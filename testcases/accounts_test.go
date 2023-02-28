package testcases

import (
	"encoding/hex"
	"github.com/deep-nl/ethgo/core"
	"strings"
	"testing"

	"github.com/deep-nl/ethgo/wallet"
	"github.com/stretchr/testify/assert"
)

func TestAccounts(t *testing.T) {
	var walletSpec []struct {
		Address    string  `json:"address"`
		Checksum   string  `json:"checksumAddress"`
		Name       string  `json:"name"`
		PrivateKey *string `json:"privateKey,omitempty"`
	}
	ReadTestCase(t, "accounts", &walletSpec)

	for _, spec := range walletSpec {
		if spec.PrivateKey != nil {
			// test that we can decode the private key
			priv, err := hex.DecodeString(strings.TrimPrefix(*spec.PrivateKey, "0x"))
			assert.NoError(t, err)

			key, err := wallet.NewWalletFromPrivKey(priv)
			assert.NoError(t, err)

			assert.Equal(t, key.Address().String(), spec.Checksum)
		}

		// test that an string address can be checksumed
		addr := core.HexToAddress(spec.Address)
		assert.Equal(t, addr.String(), spec.Checksum)
	}
}
