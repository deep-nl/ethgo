package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/deep-nl/ethgo/core"

	"github.com/btcsuite/btcd/btcec"
)

// S256 is the secp256k1 elliptic curve
var S256 = btcec.S256()

var _ core.Key = &Key{}

// Key is an implementation of the Key interface with a private key
type Key struct {
	priv *ecdsa.PrivateKey
	pub  *ecdsa.PublicKey
	addr core.Address
}

func (k *Key) Address() core.Address {
	return k.addr
}

func (k *Key) MarshallPrivateKey() ([]byte, error) {
	return (*btcec.PrivateKey)(k.priv).Serialize(), nil
}

func (k *Key) SignMsg(msg []byte) ([]byte, error) {
	return k.Sign(core.Keccak256(msg))
}

func (k *Key) Sign(hash []byte) ([]byte, error) {
	sig, err := btcec.SignCompact(S256, (*btcec.PrivateKey)(k.priv), hash, false)
	if err != nil {
		return nil, err
	}
	term := byte(0)
	if sig[0] == 28 {
		term = 1
	}
	return append(sig, term)[1:], nil
}

// NewKey creates a new key with a private key
func NewKey(priv *ecdsa.PrivateKey) *Key {
	return &Key{
		priv: priv,
		pub:  &priv.PublicKey,
		addr: pubKeyToAddress(&priv.PublicKey),
	}
}

func KeyFromString(privKeyStr string) *Key {
	// Parse the string representation of the private key
	privKeyBytes, err := hex.DecodeString(privKeyStr)
	if err != nil {
		fmt.Println("Error decoding private key:", err)
		return nil
	}

	privKeyBtcec, _ := btcec.PrivKeyFromBytes(btcec.S256(), privKeyBytes)

	// Convert the *btcec.PrivateKey to *ecdsa.PrivateKey
	privKeyEcdsa := (*ecdsa.PrivateKey)(privKeyBtcec.ToECDSA())
	return NewKey(privKeyEcdsa)
}

func pubKeyToAddress(pub *ecdsa.PublicKey) (addr core.Address) {
	b := core.Keccak256(elliptic.Marshal(S256, pub.X, pub.Y)[1:])
	copy(addr[:], b[12:])
	return
}

// GenerateKey generates a new key based on the secp256k1 elliptic curve.
func GenerateKey() (*Key, error) {
	priv, err := ecdsa.GenerateKey(S256, rand.Reader)
	if err != nil {
		return nil, err
	}
	return NewKey(priv), nil
}

func EcrecoverMsg(msg, signature []byte) (core.Address, error) {
	return Ecrecover(core.Keccak256(msg), signature)
}

func Ecrecover(hash, signature []byte) (core.Address, error) {
	pub, err := RecoverPubkey(signature, hash)
	if err != nil {
		return core.Address{}, err
	}
	return pubKeyToAddress(pub), nil
}

func RecoverPubkey(signature, hash []byte) (*ecdsa.PublicKey, error) {
	size := len(signature)
	term := byte(27)
	if signature[size-1] == 1 {
		term = 28
	}

	sig := append([]byte{term}, signature[:size-1]...)
	pub, _, err := btcec.RecoverCompact(S256, sig, hash)
	if err != nil {
		return nil, err
	}
	return pub.ToECDSA(), nil
}
