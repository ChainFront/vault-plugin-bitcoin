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
	"context"
	"testing"
	"time"

	"fmt"
	"github.com/hashicorp/vault/logical"
)

const (
	defaultLeaseTTLHr = 1
	maxLeaseTTLHr     = 12
)

// Set up/Teardown
type testData struct {
	B logical.Backend
	S logical.Storage
}

func setupTest(t *testing.T) *testData {
	b, reqStorage := getTestBackend(t)
	return &testData{
		B: b,
		S: reqStorage,
	}
}

func getTestBackend(t *testing.T) (logical.Backend, logical.Storage) {
	b := Backend()

	config := &logical.BackendConfig{
		System: &logical.StaticSystemView{
			DefaultLeaseTTLVal: defaultLeaseTTLHr * time.Hour,
			MaxLeaseTTLVal:     maxLeaseTTLHr * time.Hour,
		},
		StorageView: &logical.InmemStorage{},
	}
	err := b.Setup(context.Background(), config)
	if err != nil {
		t.Fatalf("unable to create backend: %v", err)
	}

	return b, config.StorageView
}

func TestBackend_createAccount(t *testing.T) {

	td := setupTest(t)

	accountName := "account1"

	createAccount(td, accountName, t)

	accountData := readAccount(td, accountName, t)
	address := accountData["address"].(*string)
	txSpendLimit := accountData["txSpendLimit"].(*string)
	t.Logf("successfully created account -- address:'%s', txSpendLimit:'%s'", *address, *txSpendLimit)
}

func TestBackend_signPaymentTransaction(t *testing.T) {

	td := setupTest(t)
	createAccount(td, "testSourceAccount", t)
	createAccount(td, "testDestinationAccount", t)
	unsignedTx := "01000000017b1eabe0209b1fe794124575ef807057c77ada2138ae4fa8d6c4de0398a14f3f0000000000ffffffff01f0ca052a010000001976a914cbc20a7664f2f69e5355aa427045bc15e7c6c77288ac00000000"

	respData, err := signTx(td, "testSourceAccount", "testDestinationAccount", unsignedTx, "35", t)
	if err != nil {
		t.Fatal("unable to sign transaction", err)
	}

	signedTx, ok := respData["signed_transaction"]
	if !ok {
		t.Fatalf("expected signedTx data not present in createPayment")
	}

	t.Logf("signedTx : %s", signedTx)
}

func TestBackend_signPaymentTransactionAboveLimit(t *testing.T) {

	td := setupTest(t)
	createAccount(td, "testSourceAccount", t)
	createAccount(td, "testDestinationAccount", t)
	unsignedTx := "01000000017b1eabe0209b1fe794124575ef807057c77ada2138ae4fa8d6c4de0398a14f3f0000000000ffffffff01f0ca052a010000001976a914cbc20a7664f2f69e5355aa427045bc15e7c6c77288ac00000000"

	_, err := signTx(td, "testSourceAccount", "testDestinationAccount", unsignedTx, "1001", t)

	expectedError := "transaction amount (1001) is larger than the transactional limit (1000)"
	if err.Error() != expectedError {
		t.Fatalf("limit check error not returned")
	}

}

func TestBackend_GetCurrentFee(t *testing.T) {
	fee, err := GetCurrentFeeRate()
	if err != nil {
		t.Fatalf("unable to get fee rate: %v", err)
	}
	t.Logf("current fee rate = %v", fee)
}

func createAccount(td *testData, accountName string, t *testing.T) {
	d :=
		map[string]interface{}{
			"tx_spend_limit": "1000",
		}
	resp, err := td.B.HandleRequest(context.Background(), &logical.Request{
		Operation: logical.CreateOperation,
		Path:      fmt.Sprintf("accounts/%s", accountName),
		Data:      d,
		Storage:   td.S,
	})
	if err != nil {
		t.Fatalf("failed to create account: %v", err)
	}
	if resp.IsError() {
		t.Fatal(resp.Error())
	}
	if resp == nil {
		t.Fatal("response is nil")
	}
	t.Logf("successfully created account : %v", resp.Data)
}

func readAccount(td *testData, accountName string, t *testing.T) map[string]interface{} {
	resp, err := td.B.HandleRequest(context.Background(), &logical.Request{
		Operation: logical.ReadOperation,
		Path:      fmt.Sprintf("accounts/%s", accountName),
		Storage:   td.S,
	})
	if err != nil {
		t.Fatalf("failed to read account: %v", err)
	}
	if resp.IsError() {
		t.Fatal(resp.Error())
	}
	if resp == nil {
		t.Fatal("response is nil")
	}

	return resp.Data
}

func signTx(td *testData,
	sourceAccountName string,
	destinationAccountName string,
	unsignedTx string,
	amount string,
	t *testing.T) (map[string]interface{}, error) {
	d :=
		map[string]interface{}{
			"source":      sourceAccountName,
			"destination": destinationAccountName,
			"unsignedTx":  unsignedTx,
			"amount":      amount,
		}
	resp, err := td.B.HandleRequest(context.Background(), &logical.Request{
		Operation: logical.CreateOperation,
		Path:      fmt.Sprintf("payments"),
		Data:      d,
		Storage:   td.S,
	})
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, err
	}
	if resp == nil {
		return nil, fmt.Errorf("response is nil")
	}
	t.Log(resp.Data)

	return resp.Data, nil
}
