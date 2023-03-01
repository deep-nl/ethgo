package core

import (
	"encoding/hex"
	"math/big"
)

type Wei *big.Int

//func convert(val uint64, decimals int64) *big.Int {
//	v := big.NewInt(int64(val))
//	exp := new(big.Int).Exp(big.NewInt(10), big.NewInt(decimals), nil)
//	return v.Mul(v, exp)
//}

func DecimalMul(v *big.Int, decimals int64) *big.Int {
	//v := big.NewInt(int64(val))
	exp := new(big.Int).Exp(big.NewInt(10), big.NewInt(decimals), nil)
	return v.Mul(v, exp)
}

func DecimalDiv(v *big.Int, decimals int64) *big.Int {
	//v := big.NewInt(int64(val))
	exp := new(big.Int).Exp(big.NewInt(10), big.NewInt(decimals), nil)
	return v.Div(v, exp)
}

func InEther(i *big.Int) *big.Int {
	return DecimalDiv(i, 18)
}

func ToEther(i *big.Int) *big.Int {
	j := new(big.Int).Set(i)
	return DecimalDiv(j, 18)
}

func ToFloatEther(i *big.Int) *big.Float {
	j := new(big.Float).SetInt(i)
	k := new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
	//k := new(big.Float).s
	return new(big.Float).Quo(j, k)
}

func InGwei(i *big.Int) *big.Int {
	return DecimalDiv(i, 9)
}

func ToGwei(i *big.Int) *big.Int {
	j := new(big.Int).Set(i)
	return DecimalDiv(j, 9)
}

// String2Byte32 Convert a string to a fixed length 32 byte.
func String2Byte32(s string) [32]byte {
	var res [32]byte
	decoded, err := hex.DecodeString(s)
	if err != nil {
		return res
	}
	if len(s) > 32 {
		copy(res[:], decoded)
	} else {
		copy(res[32-len(s):], decoded)
	}
	return res
}

func Byte2String(data []byte) string {
	//
	hexString := hex.EncodeToString(data)
	return hexString
}
