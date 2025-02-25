package testcases

import (
	"encoding/hex"
	"github.com/deep-nl/ethgo/core"
	"testing"

	"github.com/deep-nl/ethgo/abi"
	"github.com/deep-nl/ethgo/testutil"
	"github.com/stretchr/testify/assert"
)

func TestContract_Signatures(t *testing.T) {
	var signatures []struct {
		Name      string `json:"name"`
		Signature string `json:"signature"`
		SigHash   string `json:"sigHash"`
		Abi       string `json:"abi"`
	}
	ReadTestCase(t, "contract-signatures", &signatures)

	for _, c := range signatures {
		m, err := abi.NewMethod(c.Signature)
		assert.NoError(t, err)

		sigHash := "0x" + hex.EncodeToString(m.ID())
		assert.Equal(t, sigHash, c.SigHash)
	}
}

func TestContract_Interface(t *testing.T) {
	t.Skip()

	server := testutil.NewTestServer(t)

	var calls []struct {
		Name      string        `json:"name"`
		Interface string        `json:"interface"`
		Bytecode  core.ArgBytes `json:"bytecode"`
		Result    core.ArgBytes `json:"result"`
		Values    string        `json:"values"`
	}
	ReadTestCase(t, "contract-interface", &calls)

	for _, c := range calls {
		a, err := abi.NewABI(c.Interface)
		assert.NoError(t, err)

		method := a.GetMethod("test")

		receipt, err := server.SendTxn(&core.Transaction{
			Input: c.Bytecode.Bytes(),
		})
		assert.NoError(t, err)

		outputRaw, err := server.Call(&core.CallMsg{
			To:   &receipt.ContractAddress,
			Data: method.ID(),
		})
		assert.NoError(t, err)

		output, err := hex.DecodeString(outputRaw[2:])
		assert.NoError(t, err)

		_, err = method.Decode(output)
		assert.NoError(t, err)
	}

}
