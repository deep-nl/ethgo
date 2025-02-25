package contract

import (
	"encoding/hex"
	"github.com/deep-nl/ethgo/core"
	"math/big"
	"os"
	"testing"

	"github.com/deep-nl/ethgo/abi"
	"github.com/deep-nl/ethgo/jsonrpc"
	"github.com/deep-nl/ethgo/testutil"
	"github.com/deep-nl/ethgo/wallet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	addr0  = "0x0000000000000000000000000000000000000000"
	addr0B = core.HexToAddress(addr0)
)

func TestContract_NoInput(t *testing.T) {
	s := testutil.NewTestServer(t)
	//http.ResponseWriter()
	//http.Request{}
	//http.Handle()
	cc := &testutil.Contract{}
	cc.AddOutputCaller("set")

	contract, addr, err := s.DeployContract(cc)
	require.NoError(t, err)

	abi0, err := abi.NewABI(contract.Abi)
	assert.NoError(t, err)

	p, _ := jsonrpc.NewClient(s.HTTPAddr())
	c := NewContract(addr, abi0, WithJsonRPC(p.Eth()))

	vals, err := c.Call("set", core.Latest)
	assert.NoError(t, err)
	assert.Equal(t, vals["0"], big.NewInt(1))

	abi1, err := abi.NewABIFromList([]string{
		"function set() view returns (uint256)",
	})
	assert.NoError(t, err)

	c1 := NewContract(addr, abi1, WithJsonRPC(p.Eth()))
	vals, err = c1.Call("set", core.Latest)
	assert.NoError(t, err)
	assert.Equal(t, vals["0"], big.NewInt(1))
}

func TestContract_IO(t *testing.T) {
	s := testutil.NewTestServer(t)

	cc := &testutil.Contract{}
	cc.AddDualCaller("setA", "address", "uint256")

	contract, addr, err := s.DeployContract(cc)
	require.NoError(t, err)

	abi, err := abi.NewABI(contract.Abi)
	assert.NoError(t, err)

	c := NewContract(addr, abi, WithJsonRPCEndpoint(s.HTTPAddr()))

	resp, err := c.Call("setA", core.Latest, addr0B, 1000)
	assert.NoError(t, err)

	assert.Equal(t, resp["0"], addr0B)
	assert.Equal(t, resp["1"], big.NewInt(1000))
}

func TestContract_From(t *testing.T) {
	s := testutil.NewTestServer(t)

	cc := &testutil.Contract{}
	cc.AddCallback(func() string {
		return `function example() public view returns (address) {
			return msg.sender;	
		}`
	})

	contract, addr, err := s.DeployContract(cc)
	require.NoError(t, err)

	abi, err := abi.NewABI(contract.Abi)
	assert.NoError(t, err)

	from := core.Address{0x1}
	c := NewContract(addr, abi, WithSender(from), WithJsonRPCEndpoint(s.HTTPAddr()))

	resp, err := c.Call("example", core.Latest)
	assert.NoError(t, err)
	assert.Equal(t, resp["0"], from)
}

func TestContract_Deploy(t *testing.T) {
	s := testutil.NewTestServer(t)

	// create an address and fund it
	key, _ := wallet.GenerateKey()
	s.Fund(key.Address())

	p, _ := jsonrpc.NewClient(s.HTTPAddr())

	cc := &testutil.Contract{}
	cc.AddConstructor("address", "uint256")

	artifact, err := cc.Compile()
	assert.NoError(t, err)

	abi, err := abi.NewABI(artifact.Abi)
	assert.NoError(t, err)

	bin, err := hex.DecodeString(artifact.Bin)
	assert.NoError(t, err)

	txn, err := DeployContract(abi, bin, []interface{}{core.Address{0x1}, 1000}, WithJsonRPC(p.Eth()), WithSender(key))
	assert.NoError(t, err)

	assert.NoError(t, txn.Do())
	receipt, err := txn.Wait()
	assert.NoError(t, err)

	i := NewContract(receipt.ContractAddress, abi, WithJsonRPC(p.Eth()))
	resp, err := i.Call("val_0", core.Latest)
	assert.NoError(t, err)
	assert.Equal(t, resp["0"], core.Address{0x1})

	resp, err = i.Call("val_1", core.Latest)
	assert.NoError(t, err)
	assert.Equal(t, resp["0"], big.NewInt(1000))
}

func TestContract_Transaction(t *testing.T) {
	s := testutil.NewTestServer(t)

	// create an address and fund it
	key, _ := wallet.GenerateKey()
	s.Fund(key.Address())

	cc := &testutil.Contract{}
	cc.AddEvent(testutil.NewEvent("A").Add("uint256", true))
	cc.EmitEvent("setA", "A", "1")

	artifact, addr, err := s.DeployContract(cc)
	require.NoError(t, err)

	abi, err := abi.NewABI(artifact.Abi)
	assert.NoError(t, err)

	// send multiple transactions
	contract := NewContract(addr, abi, WithJsonRPCEndpoint(s.HTTPAddr()), WithSender(key))

	for i := 0; i < 10; i++ {

		txn, err := contract.Txn("setA")
		assert.NoError(t, err)

		err = txn.Do()
		assert.NoError(t, err)

		receipt, err := txn.Wait()
		assert.NoError(t, err)
		assert.Len(t, receipt.Logs, 1)
	}
}

func TestContract_CallAtBlock(t *testing.T) {
	s := testutil.NewTestServer(t)

	// create an address and fund it
	key, _ := wallet.GenerateKey()
	s.Fund(key.Address())

	cc := &testutil.Contract{}
	cc.AddCallback(func() string {
		return `
		uint256 val = 1;
		function getVal() public view returns (uint256) {
			return val;
		}
		function change() public payable {
			val = 2;
		}`
	})

	artifact, addr, err := s.DeployContract(cc)
	require.NoError(t, err)

	abi, err := abi.NewABI(artifact.Abi)
	assert.NoError(t, err)

	contract := NewContract(addr, abi, WithJsonRPCEndpoint(s.HTTPAddr()), WithSender(key))

	checkVal := func(block core.BlockNumber, expected *big.Int) {
		resp, err := contract.Call("getVal", block)
		assert.NoError(t, err)
		assert.Equal(t, resp["0"], expected)
	}

	// initial value is 1
	checkVal(core.Latest, big.NewInt(1))

	// send a transaction to update the state
	var receipt *core.Receipt
	{
		txn, err := contract.Txn("change")
		assert.NoError(t, err)

		err = txn.Do()
		assert.NoError(t, err)

		receipt, err = txn.Wait()
		assert.NoError(t, err)
	}

	// validate the state at different blocks
	{
		// value at receipt block is 2
		checkVal(core.BlockNumber(receipt.BlockNumber), big.NewInt(2))

		// value at previous block is 1
		checkVal(core.BlockNumber(receipt.BlockNumber-1), big.NewInt(1))
	}
}

func TestContract_SendValueContractCall(t *testing.T) {
	s := testutil.NewTestServer(t)

	key, _ := wallet.GenerateKey()
	s.Fund(key.Address())

	cc := &testutil.Contract{}
	cc.AddCallback(func() string {
		return `
		function deposit() public payable {
		}`
	})

	artifact, addr, err := s.DeployContract(cc)
	require.NoError(t, err)

	abi, err := abi.NewABI(artifact.Abi)
	assert.NoError(t, err)

	contract := NewContract(addr, abi, WithJsonRPCEndpoint(s.HTTPAddr()), WithSender(key))

	balance := big.NewInt(1)

	txn, err := contract.Txn("deposit")
	txn.WithOpts(&TxnOpts{Value: balance})
	assert.NoError(t, err)

	err = txn.Do()
	assert.NoError(t, err)

	_, err = txn.Wait()
	assert.NoError(t, err)

	client, _ := jsonrpc.NewClient(s.HTTPAddr())
	found, err := client.Eth().GetBalance(addr, core.Latest)
	assert.NoError(t, err)
	assert.Equal(t, found, balance)
}

func TestContract_EIP1559(t *testing.T) {
	s := testutil.NewTestServer(t)

	key, _ := wallet.GenerateKey()
	s.Fund(key.Address())

	cc := &testutil.Contract{}
	cc.AddOutputCaller("example")

	artifact, addr, err := s.DeployContract(cc)
	require.NoError(t, err)

	abi, err := abi.NewABI(artifact.Abi)
	assert.NoError(t, err)

	client, _ := jsonrpc.NewClient(s.HTTPAddr())
	contract := NewContract(addr, abi, WithJsonRPC(client.Eth()), WithSender(key), WithEIP1559())

	txn, err := contract.Txn("example")
	assert.NoError(t, err)

	err = txn.Do()
	assert.NoError(t, err)

	_, err = txn.Wait()
	assert.NoError(t, err)

	// get transaction from rpc endpoint
	txnObj, err := client.Eth().GetTransactionByHash(txn.Hash())
	assert.NoError(t, err)

	assert.NotZero(t, txnObj.Gas)
	assert.NotZero(t, txnObj.GasPrice)
	assert.NotZero(t, txnObj.MaxFeePerGas)
	assert.NotZero(t, txnObj.MaxPriorityFeePerGas)
}

// --------------------------------------------------------------------------------------------------------------------
// -------------------------------------------TestnetTest----------------------------------------------------------------
// --------------------------------------------------------------------------------------------------------------------

func TestContract_Basic(t *testing.T) {
	server := testutil.NewServer()
	keyFrom := os.Getenv("TEST2KEY")
	//t.Log(http, ws)
	//s := testutil.NewTestingServer(t, http, ws)

	client, _ := jsonrpc.NewClient(server.HTTPAddr())

	//key, _ := wallet.GenerateKey()
	key := wallet.KeyFromString(keyFrom)
	account, err := client.Eth().GetBalance(key.Address(), core.Latest)
	assert.NoError(t, err)
	t.Log(core.ToFloatEther(account))

	abi, err := abi.NewABIFromFile("../asset/uniswap-v2/pair.abi")
	assert.NoError(t, err)

	addr := core.HexToAddress("0x186b57aFFE222D6176347D338Ed66Ea2e20D630d") // dai_weth pair
	//contract := NewContract(addr, abi, WithJsonRPC(client.Eth()), WithSender(key), WithEIP1559())
	contract := NewContract(addr, abi, WithJsonRPC(client.Eth()))

	resp, err := contract.Call("getReserves", core.Latest)
	assert.NoError(t, err)
	t.Log(resp)

}
