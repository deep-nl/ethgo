package jsonrpc

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/deep-nl/ethgo/core"
	"math/big"
)

// Eth is the eth namespace
type Eth struct {
	c *Client
}

// Eth returns the reference to the eth namespace
func (c *Client) Eth() *Eth {
	return c.endpoints.e
}

// GetCode returns the code of a contract
func (e *Eth) GetCode(addr core.Address, block core.BlockNumberOrHash) (string, error) {
	var res string
	if err := e.c.Call("eth_getCode", &res, addr, block.Location()); err != nil {
		return "", err
	}
	return res, nil
}

// Accounts returns a list of addresses owned by client.
func (e *Eth) Accounts() ([]core.Address, error) {
	var out []core.Address
	if err := e.c.Call("eth_accounts", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetStorageAt returns the value from a storage position at a given address.
func (e *Eth) GetStorageAt(addr core.Address, slot core.Hash, block core.BlockNumberOrHash) (core.Hash, error) {
	var hash core.Hash
	err := e.c.Call("eth_getStorageAt", &hash, addr, slot, block.Location())
	return hash, err
}

// BlockNumber returns the number of most recent block.
func (e *Eth) BlockNumber() (uint64, error) {
	var out string
	if err := e.c.Call("eth_blockNumber", &out); err != nil {
		return 0, err
	}
	return parseUint64orHex(out)
}

// GetBlockByNumber returns information about a block by block number.
func (e *Eth) GetBlockByNumber(i core.BlockNumber, full bool) (*core.Block, error) {
	var b *core.Block
	if err := e.c.Call("eth_getBlockByNumber", &b, i.String(), full); err != nil {
		return nil, err
	}
	return b, nil
}

// GetBlockByHash returns information about a block by hash.
func (e *Eth) GetBlockByHash(hash core.Hash, full bool) (*core.Block, error) {
	var b *core.Block
	if err := e.c.Call("eth_getBlockByHash", &b, hash, full); err != nil {
		return nil, err
	}
	return b, nil
}

// GetFilterChanges returns the filter changes for log filters
func (e *Eth) GetFilterChanges(id string) ([]*core.Log, error) {
	var logs []*core.Log
	if err := e.c.Call("eth_getFilterChanges", &logs, id); err != nil {
		return nil, err
	}
	return logs, nil
}

// GetTransactionByHash returns a transaction by his hash
func (e *Eth) GetTransactionByHash(hash core.Hash) (*core.Transaction, error) {
	var txn *core.Transaction
	err := e.c.Call("eth_getTransactionByHash", &txn, hash)
	return txn, err
}

// GetFilterChangesBlock returns the filter changes for block filters
func (e *Eth) GetFilterChangesBlock(id string) ([]core.Hash, error) {
	var hashes []core.Hash
	if err := e.c.Call("eth_getFilterChanges", &hashes, id); err != nil {
		return nil, err
	}
	return hashes, nil
}

// NewFilter creates a new log filter
func (e *Eth) NewFilter(filter *core.LogFilter) (string, error) {
	var id string
	err := e.c.Call("eth_newFilter", &id, filter)
	return id, err
}

// NewBlockFilter creates a new block filter
func (e *Eth) NewBlockFilter() (string, error) {
	var id string
	err := e.c.Call("eth_newBlockFilter", &id, nil)
	return id, err
}

// UninstallFilter uninstalls a filter
func (e *Eth) UninstallFilter(id string) (bool, error) {
	var res bool
	err := e.c.Call("eth_uninstallFilter", &res, id)
	return res, err
}

// SendRawTransaction sends a signed transaction in rlp format.
func (e *Eth) SendRawTransaction(data []byte) (core.Hash, error) {
	var hash core.Hash
	hexData := "0x" + hex.EncodeToString(data)
	err := e.c.Call("eth_sendRawTransaction", &hash, hexData)
	return hash, err
}

// SendTransaction creates new message call transaction or a contract creation.
func (e *Eth) SendTransaction(txn *core.Transaction) (core.Hash, error) {
	var hash core.Hash
	err := e.c.Call("eth_sendTransaction", &hash, txn)
	return hash, err
}

// GetTransactionReceipt returns the receipt of a transaction by transaction hash.
func (e *Eth) GetTransactionReceipt(hash core.Hash) (*core.Receipt, error) {
	var receipt *core.Receipt
	err := e.c.Call("eth_getTransactionReceipt", &receipt, hash)
	return receipt, err
}

// GetNonce returns the nonce of the account
func (e *Eth) GetNonce(addr core.Address, blockNumber core.BlockNumberOrHash) (uint64, error) {
	var nonce string
	if err := e.c.Call("eth_getTransactionCount", &nonce, addr, blockNumber.Location()); err != nil {
		return 0, err
	}
	return parseUint64orHex(nonce)
}

// GetBalance returns the balance of the account of given address.
func (e *Eth) GetBalance(addr core.Address, blockNumber core.BlockNumberOrHash) (*big.Int, error) {
	var out string
	if err := e.c.Call("eth_getBalance", &out, addr, blockNumber.Location()); err != nil {
		return nil, err
	}
	b, ok := new(big.Int).SetString(out[2:], 16)
	if !ok {
		return nil, fmt.Errorf("failed to convert to big.int")
	}
	return b, nil
}

// GasPrice returns the current price per gas in wei.
func (e *Eth) GasPrice() (uint64, error) {
	var out string
	if err := e.c.Call("eth_gasPrice", &out); err != nil {
		return 0, err
	}
	return parseUint64orHex(out)
}

// Call executes a new message call immediately without creating a transaction on the block chain.
func (e *Eth) Call(msg *core.CallMsg, block core.BlockNumber) (string, error) {
	var out string
	if err := e.c.Call("eth_call", &out, msg, block.String()); err != nil {
		return "", err
	}
	return out, nil
}

// EstimateGasContract estimates the gas to deploy a contract
func (e *Eth) EstimateGasContract(bin []byte) (uint64, error) {
	var out string
	msg := map[string]interface{}{
		"data": "0x" + hex.EncodeToString(bin),
	}
	if err := e.c.Call("eth_estimateGas", &out, msg); err != nil {
		return 0, err
	}
	return parseUint64orHex(out)
}

// EstimateGas generates and returns an estimate of how much gas is necessary to allow the transaction to complete.
func (e *Eth) EstimateGas(msg *core.CallMsg) (uint64, error) {
	var out string
	if err := e.c.Call("eth_estimateGas", &out, msg); err != nil {
		return 0, err
	}
	return parseUint64orHex(out)
}

// GetLogs returns an array of all logs matching a given filter object
func (e *Eth) GetLogs(filter *core.LogFilter) ([]*core.Log, error) {
	var out []*core.Log
	if err := e.c.Call("eth_getLogs", &out, filter); err != nil {
		return nil, err
	}
	return out, nil
}

// ChainID returns the id of the chain
func (e *Eth) ChainID() (*big.Int, error) {
	var out string
	if err := e.c.Call("eth_chainId", &out); err != nil {
		return nil, err
	}
	return parseBigInt(out), nil
}

// FeeHistory is the result of the eth_feeHistory endpoint
type FeeHistory struct {
	OldestBlock  *big.Int     `json:"oldestBlock"`
	Reward       [][]*big.Int `json:"reward,omitempty"`
	BaseFee      []*big.Int   `json:"baseFeePerGas,omitempty"`
	GasUsedRatio []float64    `json:"gasUsedRatio"`
}

func (f *FeeHistory) UnmarshalJSON(data []byte) error {
	var raw struct {
		OldestBlock  *ArgBig     `json:"oldestBlock"`
		Reward       [][]*ArgBig `json:"reward,omitempty"`
		BaseFee      []*ArgBig   `json:"baseFeePerGas,omitempty"`
		GasUsedRatio []float64   `json:"gasUsedRatio"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if raw.OldestBlock != nil {
		f.OldestBlock = raw.OldestBlock.Big()
	}
	if raw.Reward != nil {
		f.Reward = [][]*big.Int{}
		for _, r := range raw.Reward {
			elem := []*big.Int{}
			for _, i := range r {
				elem = append(elem, i.Big())
			}
			f.Reward = append(f.Reward, elem)
		}
	}
	if raw.BaseFee != nil {
		f.BaseFee = []*big.Int{}
		for _, i := range raw.BaseFee {
			f.BaseFee = append(f.BaseFee, i.Big())
		}
	}
	f.GasUsedRatio = raw.GasUsedRatio
	return nil
}

// FeeHistory returns base fee per gas and transaction effective priority fee
func (e *Eth) FeeHistory(from, to core.BlockNumber) (*FeeHistory, error) {
	var out *FeeHistory
	if err := e.c.Call("eth_feeHistory", &out, from.String(), to.String(), nil); err != nil {
		return nil, err
	}
	return out, nil
}
