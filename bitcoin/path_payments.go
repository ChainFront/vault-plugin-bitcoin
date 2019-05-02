package bitcoin

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"github.com/btcsuite/btcd/wire"
	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
	"math/big"
)

// Register the callbacks for the paths exposed by these functions
func paymentsPaths(b *backend) []*framework.Path {
	return []*framework.Path{
		&framework.Path{
			Pattern:      "payments",
			HelpSynopsis: "Create a payment transaction",
			Fields: map[string]*framework.FieldSchema{
				"source": &framework.FieldSchema{
					Type:        framework.TypeString,
					Description: "Source account",
				},
				"destination": &framework.FieldSchema{
					Type:        framework.TypeString,
					Description: "Destination account",
				},
				"additionalSigners": &framework.FieldSchema{
					Type:        framework.TypeCommaStringSlice,
					Description: "(Optional) Array of additional signers for this transaction",
				},
				"unsignedTx": &framework.FieldSchema{
					Type:        framework.TypeString,
					Description: "The unsigned, hex-encoded transaction",
				},
				"amount": &framework.FieldSchema{
					Type:        framework.TypeString,
					Description: "Amount to send",
				},
			},
			Callbacks: map[logical.Operation]framework.OperationFunc{
				logical.CreateOperation: b.createPayment,
				logical.UpdateOperation: b.createPayment,
			},
		},
	}
}

// Bitcoin: Creates a signed transaction with a payment operation.
func (b *backend) createPayment(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {

	// Validate we didn't get extra fields
	err := validateFields(req, d)
	if err != nil {
		return nil, logical.CodedError(400, err.Error())
	}

	// Validate required fields are present
	source := d.Get("source").(string)
	if source == "" {
		return errMissingField("source"), nil
	}

	destination := d.Get("destination").(string)
	if destination == "" {
		return errMissingField("destination"), nil
	}

	unsignedTx := d.Get("unsignedTx").(string)
	if destination == "" {
		return errMissingField("unsignedTx"), nil
	}

	amountStr := d.Get("amount").(string)
	if amountStr == "" {
		return errMissingField("amount"), nil
	}
	amount := validNumber(amountStr)

	// Read the optional additionalSigners field (Note: the commercial version of this plugin uses completely
	// separate instance of Vault for addition signers)
	//var additionalSigners []string
	//if additionalSignersRaw, ok := d.GetOk("additionalSigners"); ok {
	//	additionalSigners = additionalSignersRaw.([]string)
	//}

	// Retrieve the source account keypair from vault storage
	sourceAccount, err := b.readVaultAccount(ctx, req, "accounts/"+source)
	if err != nil {
		return nil, err
	}
	if sourceAccount == nil {
		return nil, logical.CodedError(400, "source account not found")
	}
	sourceAddress := sourceAccount.Address

	// Retrieve the destination account keypair from vault storage
	destinationAccount, err := b.readVaultAccount(ctx, req, "accounts/"+destination)
	if err != nil {
		return nil, err
	}
	if destinationAccount == nil {
		return nil, logical.CodedError(400, "destination account not found")
	}

	// Validate transaction rules
	_, err = b.validAccountConstraints(sourceAccount, amount, destinationAccount.Address)
	if err != nil {
		return nil, err
	}

	// Decode the unsigned transaction
	decodedUnsignedTx, err := hex.DecodeString(unsignedTx)
	if err != nil {
		return nil, logical.CodedError(400, "unable to decode unsigned transaction: "+err.Error())
	}
	byteReader := bytes.NewReader(decodedUnsignedTx)
	tx := wire.NewMsgTx(wire.TxVersion)
	err = tx.Deserialize(byteReader)
	if err != nil {
		return nil, err
	}

	// Sign the transaction
	signedPaymentBytes, err := signPaymentTransaction(sourceAccount, tx, amount)
	if err != nil {
		return nil, logical.CodedError(500, "unable to sign transaction: "+err.Error())
	}

	signedPaymentString := hex.EncodeToString(signedPaymentBytes)

	return &logical.Response{
		Data: map[string]interface{}{
			"source_address":     sourceAddress,
			"transaction_hash":   tx.TxHash().String(),
			"signed_transaction": signedPaymentString,
		},
	}, nil
}

func (b *backend) validAccountConstraints(account *Account, amount *big.Int, toAddress string) (bool, error) {
	txLimit := validNumber(account.TxSpendLimit)

	if txLimit.Cmp(amount) == -1 && txLimit.Cmp(big.NewInt(0)) == 1 {
		return false, fmt.Errorf("transaction amount (%s) is larger than the transactional limit (%s)", amount.String(), account.TxSpendLimit)
	}

	if contains(account.Blacklist, toAddress) {
		return false, fmt.Errorf("%s is blacklisted", toAddress)
	}

	if len(account.Whitelist) > 0 && !contains(account.Whitelist, toAddress) {
		return false, fmt.Errorf("%s is not in the whitelist", toAddress)
	}

	return true, nil
}
