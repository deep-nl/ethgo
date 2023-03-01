package erc20

import (
	"github.com/deep-nl/ethgo/core"
	"testing"

	"github.com/deep-nl/ethgo/contract"
	"github.com/deep-nl/ethgo/jsonrpc"
	"github.com/deep-nl/ethgo/testutil"
	"github.com/stretchr/testify/assert"
)

var (
	zeroX = core.HexToAddress("0xe41d2489571d322189246dafa5ebde1f4699f498")

	weth = core.HexToAddress("0xB4FBF271143F4FBf7B91A5ded31805e42b2208d6")
)

func TestERC20Decimals(t *testing.T) {
	//c, _ := jsonrpc.NewClient(testutil.TestInfuraEndpoint(t))
	http := ""
	c, _ := jsonrpc.NewClient(http)

	erc20 := NewERC20(weth, contract.WithJsonRPC(c.Eth()))

	decimals, err := erc20.Decimals()
	assert.NoError(t, err)
	if decimals != 18 {
		t.Fatal("bad")
	}
}

func TestERC20Name(t *testing.T) {
	//c, _ := jsonrpc.NewClient(testutil.TestInfuraEndpoint(t))
	server := testutil.NewServer()
	c, _ := jsonrpc.NewClient(server.HTTPAddr())

	erc20 := NewERC20(weth, contract.WithJsonRPC(c.Eth()))

	name, err := erc20.Name()
	assert.NoError(t, err)
	t.Log(name)
	//assert.Equal(t, name, "0x Protocol Token")
}

func TestERC20Symbol(t *testing.T) {
	c, _ := jsonrpc.NewClient(testutil.TestInfuraEndpoint(t))
	erc20 := NewERC20(zeroX, contract.WithJsonRPC(c.Eth()))

	symbol, err := erc20.Symbol()
	assert.NoError(t, err)
	assert.Equal(t, symbol, "ZRX")
}

func TestTotalSupply(t *testing.T) {
	c, _ := jsonrpc.NewClient(testutil.TestInfuraEndpoint(t))
	erc20 := NewERC20(zeroX, contract.WithJsonRPC(c.Eth()))

	supply, err := erc20.TotalSupply()
	assert.NoError(t, err)
	assert.Equal(t, supply.String(), "1000000000000000000000000000")
}
