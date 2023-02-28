package contract

import (
	"fmt"
	"github.com/deep-nl/ethgo/core"
	"github.com/deep-nl/ethgo/jsonrpc"
	"github.com/deep-nl/ethgo/wallet"
	"math/big"
)

type jsonrpcTransaction struct {
	to      core.Address
	input   []byte
	hash    core.Hash
	opts    *TxnOpts
	key     core.Key
	client  *jsonrpc.Eth
	txn     *core.Transaction
	txnRaw  []byte
	eip1559 bool
}

func (j *jsonrpcTransaction) Hash() core.Hash {
	return j.hash
}

func (j *jsonrpcTransaction) WithOpts(opts *TxnOpts) {
	j.opts = opts
}

func (j *jsonrpcTransaction) Build() error {
	var err error
	from := j.key.Address()

	// estimate gas price
	if j.opts.GasPrice == 0 && !j.eip1559 {
		j.opts.GasPrice, err = j.client.GasPrice()
		if err != nil {
			return err
		}
	}
	// estimate gas limit
	if j.opts.GasLimit == 0 {
		msg := &core.CallMsg{
			From:     from,
			To:       nil,
			Data:     j.input,
			Value:    j.opts.Value,
			GasPrice: j.opts.GasPrice,
		}
		if j.to != core.ZeroAddress {
			msg.To = &j.to
		}
		j.opts.GasLimit, err = j.client.EstimateGas(msg)
		if err != nil {
			return err
		}
	}
	// calculate the nonce
	if j.opts.Nonce == 0 {
		j.opts.Nonce, err = j.client.GetNonce(from, core.Latest)
		if err != nil {
			return fmt.Errorf("failed to calculate nonce: %v", err)
		}
	}

	chainID, err := j.client.ChainID()
	if err != nil {
		return err
	}

	// send transaction
	rawTxn := &core.Transaction{
		From:     from,
		Input:    j.input,
		GasPrice: j.opts.GasPrice,
		Gas:      j.opts.GasLimit,
		Value:    j.opts.Value,
		Nonce:    j.opts.Nonce,
		ChainID:  chainID,
	}
	if j.to != core.ZeroAddress {
		rawTxn.To = &j.to
	}

	if j.eip1559 {
		rawTxn.Type = core.TransactionDynamicFee

		// use gas price as fee data
		gasPrice, err := j.client.GasPrice()
		if err != nil {
			return err
		}
		rawTxn.MaxFeePerGas = new(big.Int).SetUint64(gasPrice)
		rawTxn.MaxPriorityFeePerGas = new(big.Int).SetUint64(gasPrice)
	}

	j.txn = rawTxn
	return nil
}

func (j *jsonrpcTransaction) Do() error {
	if j.txn == nil {
		if err := j.Build(); err != nil {
			return err
		}
	}

	signer := wallet.NewEIP155Signer(j.txn.ChainID.Uint64())
	signedTxn, err := signer.SignTx(j.txn, j.key)
	if err != nil {
		return err
	}
	txnRaw, err := signedTxn.MarshalRLPTo(nil)
	if err != nil {
		return err
	}

	j.txnRaw = txnRaw
	hash, err := j.client.SendRawTransaction(j.txnRaw)
	if err != nil {
		return err
	}
	j.hash = hash
	return nil
}

func (j *jsonrpcTransaction) Wait() (*core.Receipt, error) {
	if (j.hash == core.Hash{}) {
		panic("transaction not executed")
	}

	for {
		receipt, err := j.client.GetTransactionReceipt(j.hash)
		if err != nil {
			if err.Error() != "not found" {
				return nil, err
			}
		}
		if receipt != nil {
			return receipt, nil
		}
	}
}
