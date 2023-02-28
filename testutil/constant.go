package testutil

import (
	"github.com/deep-nl/ethgo/core"
)

var (
	DefaultGasPrice = uint64(1879048192) // 0x70000000
	DefaultGasLimit = uint64(5242880)    // 0x500000
)

var (
	DummyAddr = core.HexToAddress("0x015f68893a39b3ba0681584387670ff8b00f4db2")
)

const (
	FromKey  string = "0a9b29e781f64c48a38a46d7a29b78f3c3a3b380e573cb50e2e917d5e7c3be3a"
	FromAddr string = "0x3A92a507acD85dcBE848680eD9D2340DB70be2d0"
	ToAddr   string = "0xdb04bd70d2D834D6845D515278E4De1F1Cba4572"
	Value    int64  = 1e18
)
