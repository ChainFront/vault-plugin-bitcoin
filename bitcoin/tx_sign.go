/*
 * Copyright (c) 2019 ChainFront LLC.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package bitcoin

import (
	"bytes"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/pkg/errors"
	"log"
	"math/big"
)

// Sign a payment transaction
func signPaymentTransaction(account *Account, paymentTx *wire.MsgTx, amount *big.Int) ([]byte, error) {

	wifEncoded := account.WIF
	wif, err := btcutil.DecodeWIF(wifEncoded)
	if err != nil {
		return nil, err
	}

	privateKey := wif.PrivKey

	sourceAddress, err := btcutil.DecodeAddress(account.Address, &chaincfg.TestNet3Params)
	if err != nil {
		return nil, err
	}

	sourcePkScript, err := txscript.PayToAddrScript(sourceAddress)
	if err != nil {
		return nil, err
	}

	sourceTxs := paymentTx.TxIn
	for i := range sourceTxs {
		sigScript, err := txscript.SignatureScript(paymentTx, i, sourcePkScript, txscript.SigHashAll, privateKey, true)
		if err != nil {
			return nil, errors.Errorf("could not generate pubSig; err: %v", err)
		}

		sourceTx := paymentTx.TxIn[i]
		sourceTx.SignatureScript = sigScript

		// TODO: Handle segwit... some POC code is below
		//paymentTxSigHashes := txscript.NewTxSigHashes(paymentTx)
		//wit, err := txscript.WitnessSignature(paymentTx, paymentTxSigHashes, i, amount, subscript, txscript.SigHashAll, privateKey, true)
		//if err != nil {
		//	return nil, errors.Errorf("could not generate witnessSig; err: %v", err)
		//}
		//paymentTx.TxIn[i].Witness = wit

		// Prove that the transaction has been validly signed by executing the script pair.
		flags := txscript.StandardVerifyFlags
		vm, err := txscript.NewEngine(sourcePkScript, paymentTx, i, flags, nil, nil, amount.Int64())
		if err != nil {
			return nil, err
		}
		if err := vm.Execute(); err != nil {
			return nil, err
		}

		log.Println("transaction successfully signed")
	}

	buf := bytes.NewBuffer(make([]byte, 0, paymentTx.SerializeSize()))
	paymentTx.Serialize(buf)

	return buf.Bytes(), nil
}
