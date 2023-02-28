package jsonrpc

import (
	"bytes"
	"encoding/hex"
	"github.com/deep-nl/ethgo/core"
	"github.com/deep-nl/ethgo/wallet"
	"math/big"
	"strings"
	"testing"

	"github.com/deep-nl/ethgo/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	addr0 = core.Address{0x1}
	addr1 = core.Address{0x2}
)

func TestEthAccounts(t *testing.T) {
	testutil.MultiAddr(t, func(s *testutil.TestServer, addr string) {
		c, _ := NewClient(addr)
		defer c.Close()

		_, err := c.Eth().Accounts()
		assert.NoError(t, err)
	})
}

func TestEthBlockNumber(t *testing.T) {
	testutil.MultiAddr(t, func(s *testutil.TestServer, addr string) {
		c, _ := NewClient(addr)
		defer c.Close()

		num, err := c.Eth().BlockNumber()
		require.NoError(t, err)

		for i := 0; i < 10; i++ {
			require.NoError(t, s.ProcessBlock())

			// since it is concurrent, we cannot ensure sequential numbers
			newNum, err := c.Eth().BlockNumber()
			require.NoError(t, err)
			require.Greater(t, newNum, num)

			num = newNum
		}
	})
}

func TestEthGetCode(t *testing.T) {
	s := testutil.NewTestServer(t)

	c, _ := NewClient(s.HTTPAddr())

	cc := &testutil.Contract{}
	cc.AddEvent(testutil.NewEvent("A").
		Add("address", true).
		Add("address", true))

	cc.EmitEvent("setA1", "A", addr0.String(), addr1.String())
	cc.EmitEvent("setA2", "A", addr1.String(), addr0.String())

	_, addr, err := s.DeployContract(cc)
	require.NoError(t, err)

	code, err := c.Eth().GetCode(addr, core.Latest)
	assert.NoError(t, err)
	assert.NotEqual(t, code, "0x")

	code2, err := c.Eth().GetCode(addr, core.BlockNumber(0))
	assert.NoError(t, err)
	assert.Equal(t, code2, "0x")
}

func TestEthGetBalance(t *testing.T) {
	s := testutil.NewTestServer(t)

	c, _ := NewClient(s.HTTPAddr())

	balance, err := c.Eth().GetBalance(s.Account(0), core.Latest)
	assert.NoError(t, err)
	assert.NotEqual(t, balance, big.NewInt(0))

	balance, err = c.Eth().GetBalance(core.Address{}, core.Latest)
	assert.NoError(t, err)
	assert.Equal(t, balance, big.NewInt(0))
}

func TestEthGetBlockByNumber(t *testing.T) {
	s := testutil.NewTestServer(t)

	c, _ := NewClient(s.HTTPAddr())

	block, err := c.Eth().GetBlockByNumber(0, true)
	assert.NoError(t, err)
	assert.Equal(t, block.Number, uint64(0))

	// query a non-sealed block block 1 has not been processed yet
	// it does not fail but returns nil
	latest, err := c.Eth().BlockNumber()
	require.NoError(t, err)

	block, err = c.Eth().GetBlockByNumber(core.BlockNumber(latest+10000), true)
	assert.NoError(t, err)
	assert.Nil(t, block)
}

func TestEthGetBlockByHash(t *testing.T) {
	testutil.MultiAddr(t, func(s *testutil.TestServer, addr string) {
		c, _ := NewClient(addr)
		defer c.Close()

		// get block 0 first by number
		block, err := c.Eth().GetBlockByNumber(0, true)
		assert.NoError(t, err)
		assert.Equal(t, block.Number, uint64(0))

		// get block 0 by hash
		block2, err := c.Eth().GetBlockByHash(block.Hash, true)
		assert.NoError(t, err)
		assert.Equal(t, block, block2)
	})
}

func TestEthGasPrice(t *testing.T) {
	testutil.MultiAddr(t, func(s *testutil.TestServer, addr string) {
		c, _ := NewClient(addr)
		defer c.Close()

		_, err := c.Eth().GasPrice()
		assert.NoError(t, err)
	})
}

func TestEthSendTransaction(t *testing.T) {
	s := testutil.NewTestServer(t)

	c, _ := NewClient(s.HTTPAddr())

	txn := &core.Transaction{
		From:     s.Account(0),
		GasPrice: testutil.DefaultGasPrice,
		Gas:      testutil.DefaultGasLimit,
		To:       &testutil.DummyAddr,
		Value:    big.NewInt(10),
		Nonce:    core.Local,
	}
	hash, err := c.Eth().SendTransaction(txn)
	assert.NoError(t, err)

	var receipt *core.Receipt
	for {
		receipt, err = c.Eth().GetTransactionReceipt(hash)
		if err != nil {
			t.Fatal(err)
		}
		if receipt != nil {
			break
		}
	}
}

func TestEth_SendRawTransaction(t *testing.T) {
	s := testutil.NewTestServer(t)

	c, _ := NewClient(s.HTTPAddr())
	//c.SetMaxConnsLimit(0)
	//toAddr := ethgo.HexToAddress(testutil.ToAddr)
	txn := &core.Transaction{
		From:     s.Account(0),
		GasPrice: testutil.DefaultGasPrice,
		Gas:      testutil.DefaultGasLimit,
		To:       &testutil.DummyAddr,
		Value:    big.NewInt(testutil.Value),
	}

	fromKey := testutil.FromKey
	key := wallet.KeyFromString(fromKey)

	signer := wallet.NewEIP155Signer(core.Local)
	txn, err := signer.SignTx(txn, key)
	assert.NoError(t, err)
	data, err := txn.MarshalRLPTo(nil)
	assert.NoError(t, err)
	hash, err := c.Eth().SendRawTransaction(data)
	assert.NoError(t, err)
	t.Logf("hash: %v", hash)

	_, err = s.WaitForReceipt(hash)
	assert.NoError(t, err)

	balance, err := c.Eth().GetBalance(testutil.DummyAddr, core.Latest)
	assert.NoError(t, err)
	assert.Equal(t, balance, txn.Value)

}

func TestEthEstimateGas(t *testing.T) {
	s := testutil.NewTestServer(t)

	c, _ := NewClient(s.HTTPAddr())

	cc := &testutil.Contract{}
	cc.AddEvent(testutil.NewEvent("A").Add("address", true))
	cc.EmitEvent("setA", "A", addr0.String())

	// estimate gas to deploy the contract
	solcContract, err := cc.Compile()
	assert.NoError(t, err)

	input, err := hex.DecodeString(solcContract.Bin)
	assert.NoError(t, err)

	gas, err := c.Eth().EstimateGasContract(input)
	assert.NoError(t, err)
	assert.Greater(t, gas, uint64(140000))

	_, addr, err := s.DeployContract(cc)
	require.NoError(t, err)

	msg := &core.CallMsg{
		From: s.Account(0),
		To:   &addr,
		Data: testutil.MethodSig("setA"),
	}

	gas, err = c.Eth().EstimateGas(msg)
	assert.NoError(t, err)
	assert.NotEqual(t, gas, 0)
}

func TestEthGetLogs(t *testing.T) {
	s := testutil.NewTestServer(t)

	c, _ := NewClient(s.HTTPAddr())

	cc := &testutil.Contract{}
	cc.AddEvent(testutil.NewEvent("A").
		Add("address", true).
		Add("address", true))

	cc.EmitEvent("setA1", "A", addr0.String(), addr1.String())
	cc.EmitEvent("setA2", "A", addr1.String(), addr0.String())

	_, addr, err := s.DeployContract(cc)
	require.NoError(t, err)

	r, err := s.TxnTo(addr, "setA2")
	require.NoError(t, err)

	filter := &core.LogFilter{
		BlockHash: &r.BlockHash,
	}
	logs, err := c.Eth().GetLogs(filter)
	assert.NoError(t, err)
	assert.Len(t, logs, 1)

	log := logs[0]
	assert.Len(t, log.Topics, 3)
	assert.Equal(t, log.Address, addr)

	// first topic is the signature of the event
	assert.Equal(t, log.Topics[0].String(), cc.GetEvent("A").Sig())

	// topics have 32 bytes and the addr are 20 bytes, then, assert.Equal wont work.
	// this is a workaround until we build some helper function to test this better
	assert.True(t, bytes.HasSuffix(log.Topics[1][:], addr1[:]))
	assert.True(t, bytes.HasSuffix(log.Topics[2][:], addr0[:]))
}

func TestEthChainID(t *testing.T) {
	testutil.MultiAddr(t, func(s *testutil.TestServer, addr string) {
		c, _ := NewClient(addr)
		defer c.Close()

		num, err := c.Eth().ChainID()
		assert.NoError(t, err)
		assert.Equal(t, num.Uint64(), uint64(1337)) // chainid of geth-dev
	})
}

func TestEthGetNonce(t *testing.T) {
	s := testutil.NewTestServer(t)

	c, _ := NewClient(s.HTTPAddr())

	//receipt, err := s.ProcessBlockWithReceipt()
	receipt, err := s.ProcessRawTxWithReceipt()
	assert.NoError(t, err)

	// query the balance with different options
	cases := []core.BlockNumberOrHash{
		core.Latest,
		receipt.BlockHash,
		core.BlockNumber(receipt.BlockNumber),
	}
	for _, ca := range cases {
		t.Logf("Block: %v", ca)
		num, err := c.Eth().GetNonce(s.Account(0), ca)
		assert.NoError(t, err)
		assert.NotEqual(t, num, uint64(0))
	}
}

func TestEthTransactionsInBlock(t *testing.T) {
	s := testutil.NewTestServer(t)

	c, _ := NewClient(s.HTTPAddr())

	// block 0 does not have transactions
	_, err := c.Eth().GetBlockByNumber(0, false)
	assert.NoError(t, err)

	// Process a block with a transaction
	//assert.NoError(t, s.ProcessBlock())
	assert.NoError(t, s.ProcessBlockRaw())

	latest, err := c.Eth().BlockNumber()
	require.NoError(t, err)

	num := core.BlockNumber(latest)

	// get a non-full block
	block0, err := c.Eth().GetBlockByNumber(num, false)
	assert.NoError(t, err)

	assert.NotEmpty(t, block0.TransactionsHashes, 1)
	assert.Empty(t, block0.Transactions, 0)

	// get a full block
	block1, err := c.Eth().GetBlockByNumber(num, true)
	assert.NoError(t, err)

	assert.Empty(t, block1.TransactionsHashes, 0)
	assert.NotEmpty(t, block1.Transactions, 1)

	for indx := range block0.TransactionsHashes {
		assert.Equal(t, block0.TransactionsHashes[indx], block1.Transactions[indx].Hash)
	}
}

func TestEthGetStorageAt(t *testing.T) {
	s := testutil.NewTestServer(t)

	c, _ := NewClient(s.HTTPAddr())

	cc := &testutil.Contract{}

	// add global variables
	cc.AddCallback(func() string {
		return "uint256 val;"
	})

	// add setter method
	cc.AddCallback(func() string {
		return `function setValue() public payable {
			val = 10;
		}`
	})

	_, addr, err := s.DeployContract(cc)
	require.NoError(t, err)

	receipt, err := s.TxnTo(addr, "setValue")
	require.NoError(t, err)

	cases := []core.BlockNumberOrHash{
		core.Latest,
		receipt.BlockHash,
		core.BlockNumber(receipt.BlockNumber),
	}
	for _, ca := range cases {
		res, err := c.Eth().GetStorageAt(addr, core.Hash{}, ca)
		assert.NoError(t, err)
		assert.True(t, strings.HasSuffix(res.String(), "a"))
	}
}

func TestEthFeeHistory(t *testing.T) {
	url := "https://goerli.infura.io/v3/c436c6349f034f7ba79623c7e6fe4014"
	//c, _ := NewClient(testutil.TestInfuraEndpoint(t))
	c, _ := NewClient(url)

	lastBlock, err := c.Eth().BlockNumber()
	assert.NoError(t, err)

	from := core.BlockNumber(lastBlock - 2)
	to := core.BlockNumber(lastBlock)

	fee, err := c.Eth().FeeHistory(from, to)
	assert.NoError(t, err)
	assert.NotNil(t, fee)
}

// --------------------------------------------------------------------------------------------------------------------
// -------------------------------------------NewTestingServer----------------------------------------------------------------
// --------------------------------------------------------------------------------------------------------------------

func TestEth_SendRawTransaction_withNewServer(t *testing.T) {
	s := testutil.NewTestingServer(t)

	//c := s.HttpClient()
	//c, _ := NewClient(s.HTTPAddr())
	//c.SetMaxConnsLimit(0)
	//toAddr := ethgo.HexToAddress(testutil.ToAddr)
	txn := &core.Transaction{
		From:     s.Account(0),
		GasPrice: testutil.DefaultGasPrice,
		Gas:      testutil.DefaultGasLimit,
		To:       &testutil.DummyAddr,
		Value:    big.NewInt(testutil.Value),
	}
	rec, err := s.SendRawTxn(testutil.FromKey, txn)
	assert.NoError(t, err)
	hash := rec.TransactionHash
	t.Logf("hash: %v", hash)

	_, err = s.WaitForReceipt(hash)
	assert.NoError(t, err)
}

func Test_MaxListenerNum(t *testing.T) {
	for i := 0; i < 15; i++ {
		s := testutil.NewServer()
		s.ProcessBlockRaw()

	}
}
