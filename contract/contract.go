package contract

import (
	"encoding/hex"
	"fmt"
	"github.com/deep-nl/ethgo/core"
	"math/big"

	"github.com/deep-nl/ethgo/abi"
	"github.com/deep-nl/ethgo/jsonrpc"
)

// Provider handles the interactions with the Ethereum 1x node
type Provider interface {
	Call(core.Address, []byte, *CallOpts) ([]byte, error)
	Txn(core.Address, core.Key, []byte) (Txn, error)
}

type jsonRPCNodeProvider struct {
	client  *jsonrpc.Eth
	eip1559 bool
}

func (j *jsonRPCNodeProvider) Call(addr core.Address, input []byte, opts *CallOpts) ([]byte, error) {
	msg := &core.CallMsg{
		To:   &addr,
		Data: input,
	}
	if opts.From != core.ZeroAddress {
		msg.From = opts.From
	}
	rawStr, err := j.client.Call(msg, opts.Block)
	if err != nil {
		return nil, err
	}
	raw, err := hex.DecodeString(rawStr[2:])
	if err != nil {
		return nil, err
	}
	return raw, nil
}

func (j *jsonRPCNodeProvider) Txn(addr core.Address, key core.Key, input []byte) (Txn, error) {
	txn := &jsonrpcTransaction{
		opts:    &TxnOpts{},
		input:   input,
		client:  j.client,
		key:     key,
		to:      addr,
		eip1559: j.eip1559,
	}
	return txn, nil
}

// Txn is the transaction object returned
type Txn interface {
	Hash() core.Hash
	WithOpts(opts *TxnOpts)
	Do() error
	Wait() (*core.Receipt, error)
}

type Opts struct {
	JsonRPCEndpoint string
	JsonRPCClient   *jsonrpc.Eth
	Provider        Provider
	Sender          core.Key
	EIP1559         bool
}

type ContractOption func(*Opts)

func WithJsonRPCEndpoint(endpoint string) ContractOption {
	return func(o *Opts) {
		o.JsonRPCEndpoint = endpoint
	}
}

func WithJsonRPC(client *jsonrpc.Eth) ContractOption {
	return func(o *Opts) {
		o.JsonRPCClient = client
	}
}

func WithProvider(provider Provider) ContractOption {
	return func(o *Opts) {
		o.Provider = provider
	}
}

func WithSender(sender core.Key) ContractOption {
	return func(o *Opts) {
		o.Sender = sender
	}
}

func WithEIP1559() ContractOption {
	return func(o *Opts) {
		o.EIP1559 = true
	}
}

func DeployContract(abi *abi.ABI, bin []byte, args []interface{}, opts ...ContractOption) (Txn, error) {
	a := NewContract(core.Address{}, abi, opts...)
	a.bin = bin
	return a.Txn("constructor", args...)
}

func NewContract(addr core.Address, abi *abi.ABI, opts ...ContractOption) *Contract {
	opt := &Opts{
		JsonRPCEndpoint: "http://localhost:8545",
	}
	for _, c := range opts {
		c(opt)
	}

	var provider Provider
	if opt.Provider != nil {
		provider = opt.Provider
	} else if opt.JsonRPCClient != nil {
		provider = &jsonRPCNodeProvider{client: opt.JsonRPCClient, eip1559: opt.EIP1559}
	} else {
		client, _ := jsonrpc.NewClient(opt.JsonRPCEndpoint)
		provider = &jsonRPCNodeProvider{client: client.Eth(), eip1559: opt.EIP1559}
	}

	a := &Contract{
		addr:     addr,
		abi:      abi,
		provider: provider,
		key:      opt.Sender,
	}

	return a
}

// Contract is a wrapper to make abi calls to contract with a state provider
type Contract struct {
	addr     core.Address
	abi      *abi.ABI
	bin      []byte
	provider Provider
	key      core.Key
}

func (a *Contract) GetABI() *abi.ABI {
	return a.abi
}

type TxnOpts struct {
	Value    *big.Int
	GasPrice uint64
	GasLimit uint64
	Nonce    uint64
}

func (a *Contract) Txn(method string, args ...interface{}) (Txn, error) {
	if a.key == nil {
		return nil, fmt.Errorf("no key selected")
	}

	isContractDeployment := method == "constructor"

	var input []byte
	if isContractDeployment {
		input = append(input, a.bin...)
	}

	var abiMethod *abi.Method
	if isContractDeployment {
		if a.abi.Constructor != nil {
			abiMethod = a.abi.Constructor
		}
	} else {
		if abiMethod = a.abi.GetMethod(method); abiMethod == nil {
			return nil, fmt.Errorf("method %s not found", method)
		}
	}
	if abiMethod != nil {
		data, err := abi.Encode(args, abiMethod.Inputs)
		if err != nil {
			return nil, fmt.Errorf("failed to encode arguments: %v", err)
		}
		if isContractDeployment {
			input = append(input, data...)
		} else {
			input = append(abiMethod.ID(), data...)
		}
	}

	txn, err := a.provider.Txn(a.addr, a.key, input)
	if err != nil {
		return nil, err
	}
	return txn, nil
}

type CallOpts struct {
	Block core.BlockNumber
	From  core.Address
}

func (a *Contract) Call(method string, block core.BlockNumber, args ...interface{}) (map[string]interface{}, error) {
	m := a.abi.GetMethod(method)
	if m == nil {
		return nil, fmt.Errorf("method %s not found", method)
	}

	data, err := m.Encode(args)
	if err != nil {
		return nil, err
	}

	opts := &CallOpts{
		Block: block,
	}
	if a.key != nil {
		opts.From = a.key.Address()
	}
	rawOutput, err := a.provider.Call(a.addr, data, opts)
	if err != nil {
		return nil, err
	}

	resp, err := m.Decode(rawOutput)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
