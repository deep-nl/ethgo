package e2e

import (
	"github.com/deep-nl/ethgo/core"
	"math/big"
	"testing"

	"github.com/deep-nl/ethgo/jsonrpc"
	"github.com/deep-nl/ethgo/testutil"
	"github.com/deep-nl/ethgo/wallet"
	"github.com/stretchr/testify/assert"
)

func TestSendSignedTransaction(t *testing.T) {
	s := testutil.NewTestServer(t)

	key, err := wallet.GenerateKey()
	assert.NoError(t, err)

	// add value to the new key
	value := big.NewInt(1000000000000000000)
	s.Transfer(key.Address(), value)

	c, _ := jsonrpc.NewClient(s.HTTPAddr())

	found, _ := c.Eth().GetBalance(key.Address(), core.Latest)
	assert.Equal(t, found, value)

	chainID, err := c.Eth().ChainID()
	assert.NoError(t, err)

	// send a signed transaction
	to := core.Address{0x1}
	transferVal := big.NewInt(1000)

	gasPrice, err := c.Eth().GasPrice()
	assert.NoError(t, err)

	txn := &core.Transaction{
		To:       &to,
		Value:    transferVal,
		Nonce:    0,
		GasPrice: gasPrice,
	}

	{
		msg := &core.CallMsg{
			From:     key.Address(),
			To:       &to,
			Value:    transferVal,
			GasPrice: gasPrice,
		}
		limit, err := c.Eth().EstimateGas(msg)
		assert.NoError(t, err)

		txn.Gas = limit
	}

	// creat a signer to signature the txn
	signer := wallet.NewEIP155Signer(chainID.Uint64())
	txn, err = signer.SignTx(txn, key)
	assert.NoError(t, err)

	from, err := signer.RecoverSender(txn)
	assert.NoError(t, err)
	assert.Equal(t, from, key.Address())

	data, err := txn.MarshalRLPTo(nil)
	assert.NoError(t, err)

	hash, err := c.Eth().SendRawTransaction(data)
	assert.NoError(t, err)

	_, err = s.WaitForReceipt(hash)
	assert.NoError(t, err)

	balance, err := c.Eth().GetBalance(to, core.Latest)
	assert.NoError(t, err)
	assert.Equal(t, balance, transferVal)
}
