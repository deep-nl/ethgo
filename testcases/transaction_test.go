package testcases

import (
	"github.com/deep-nl/ethgo/core"
	"math/big"
	"testing"

	"github.com/deep-nl/ethgo/wallet"
	"github.com/stretchr/testify/assert"
)

func getUint64FromBigInt(b *core.ArgBig) (uint64, bool) {
	g := (*big.Int)(b)
	if !g.IsUint64() {
		return 0, false
	}
	return g.Uint64(), true
}

func TestTransactions(t *testing.T) {
	var transactions []struct {
		Name              string        `json:"name"`
		AccountAddress    core.Address  `json:"accountAddress"`
		PrivateKey        core.ArgBytes `json:"privateKey"`
		SignedTransaction core.ArgBytes `json:"signedTransactionChainId5"`

		Data     *core.ArgBytes  `json:"data,omitempty"`
		Value    *core.ArgBig    `json:"value,omitempty"`
		To       *core.Address   `json:"to,omitempty"`
		GasLimit *core.ArgBig    `json:"gasLimit,omitempty"`
		Nonce    *core.ArgUint64 `json:"nonce,omitempty"`
		GasPrice *core.ArgBig    `json:"gasPrice,omitempty"`
	}
	ReadTestCase(t, "transactions", &transactions)

	for _, c := range transactions {
		key, err := wallet.NewWalletFromPrivKey(c.PrivateKey)
		assert.NoError(t, err)
		assert.Equal(t, key.Address(), c.AccountAddress)

		txn := &core.Transaction{
			ChainID: big.NewInt(5),
		}
		if c.Data != nil {
			txn.Input = *c.Data
		}
		if c.Value != nil {
			txn.Value = (*big.Int)(c.Value)
		}
		if c.To != nil {
			txn.To = c.To
		}
		if c.GasLimit != nil {
			gasLimit, isUint64 := getUint64FromBigInt(c.GasLimit)
			if !isUint64 {
				return
			}
			txn.Gas = gasLimit
		}
		if c.Nonce != nil {
			txn.Nonce = c.Nonce.Uint64()
		}
		if c.GasPrice != nil {
			gasPrice, isUint64 := getUint64FromBigInt(c.GasPrice)
			if !isUint64 {
				return
			}
			txn.GasPrice = gasPrice
		}

		signer := wallet.NewEIP155Signer(5)
		signedTxn, err := signer.SignTx(txn, key)
		assert.NoError(t, err)

		txnRaw, err := signedTxn.MarshalRLPTo(nil)
		assert.NoError(t, err)
		assert.Equal(t, txnRaw, c.SignedTransaction.Bytes())
	}
}

func TestTypedTransactions(t *testing.T) {
	var transactions []struct {
		Name           string        `json:"name"`
		AccountAddress core.Address  `json:"address"`
		Key            core.ArgBytes `json:"key"`
		Signed         core.ArgBytes `json:"signed"`

		Tx struct {
			Type                 core.TransactionType
			Data                 *core.ArgBytes  `json:"data,omitempty"`
			GasLimit             *core.ArgBig    `json:"gasLimit,omitempty"`
			MaxPriorityFeePerGas *core.ArgBig    `json:"maxPriorityFeePerGas,omitempty"`
			MaxFeePerGas         *core.ArgBig    `json:"maxFeePerGas,omitempty"`
			Nonce                uint64          `json:"nonce,omitempty"`
			To                   *core.Address   `json:"to,omitempty"`
			Value                *core.ArgBig    `json:"value,omitempty"`
			GasPrice             *core.ArgBig    `json:"gasPrice,omitempty"`
			ChainID              uint64          `json:"chainId,omitempty"`
			AccessList           core.AccessList `json:"accessList,omitempty"`
		}
	}
	ReadTestCase(t, "typed-transactions", &transactions)

	for _, c := range transactions {
		key, err := wallet.NewWalletFromPrivKey(c.Key)
		assert.NoError(t, err)
		assert.Equal(t, key.Address(), c.AccountAddress)

		chainID := big.NewInt(int64(c.Tx.ChainID))

		txn := &core.Transaction{
			ChainID:              chainID,
			Type:                 c.Tx.Type,
			MaxPriorityFeePerGas: (*big.Int)(c.Tx.MaxPriorityFeePerGas),
			MaxFeePerGas:         (*big.Int)(c.Tx.MaxFeePerGas),
			AccessList:           c.Tx.AccessList,
		}
		if c.Tx.Data != nil {
			txn.Input = *c.Tx.Data
		}
		if c.Tx.Value != nil {
			txn.Value = (*big.Int)(c.Tx.Value)
		}
		if c.Tx.To != nil {
			txn.To = c.Tx.To
		}
		if c.Tx.GasLimit != nil {
			gasLimit, isUint64 := getUint64FromBigInt(c.Tx.GasLimit)
			if !isUint64 {
				return
			}
			txn.Gas = gasLimit
		}
		txn.Nonce = c.Tx.Nonce
		if c.Tx.GasPrice != nil {
			gasPrice, isUint64 := getUint64FromBigInt(c.Tx.GasPrice)
			if !isUint64 {
				return
			}
			txn.GasPrice = gasPrice
		}

		signer := wallet.NewEIP155Signer(chainID.Uint64())
		signedTxn, err := signer.SignTx(txn, key)
		assert.NoError(t, err)

		txnRaw, err := signedTxn.MarshalRLPTo(nil)
		assert.NoError(t, err)

		assert.Equal(t, txnRaw, c.Signed.Bytes())
	}
}
