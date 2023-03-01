package testutil

import (
	"encoding/hex"
	"fmt"
	"github.com/deep-nl/ethgo/core"
	"github.com/deep-nl/ethgo/wallet"
	"github.com/joho/godotenv"
	"log"
	"math/big"
	"os"
	"path"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/deep-nl/ethgo/compiler"
)

// ServerConfig is the configuration of the server
type ServerConfig struct {
	Period int
}

// Server is a Geth test server
type Server struct {
	httpUrl    string
	wsUrl      string
	accounts   []core.Address
	httpClient *ethClient
	wsClient   *ethClient
}

// NewTestingServer just for testing
func NewTestingServer(t *testing.T, addrs ...string) *Server {
	if len(addrs) == 0 {
		addrs = []string{"http://127.0.0.1:8545", "ws://127.0.0.1:8545"}
	}
	server := &Server{}
	for _, url := range addrs {
		if strings.HasPrefix(url, "http") {
			server.httpUrl = url
			server.httpClient = &ethClient{url}
		} else if strings.HasPrefix(url, "ws") {
			server.wsUrl = url
			server.wsClient = &ethClient{url}
		} else {
			t.Fatal("Incorrect url format")
		}
	}

	if strings.HasSuffix(server.httpUrl, "8545") {
		t.Log("Fetch default account")
		server.accounts = append(server.accounts, core.HexToAddress(FromAddr))
		return server
	}
	return server
}

func NewServer(addrs ...string) *Server {
	// ------------------------------extract pat --------
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("Failed to get current directory")
	}
	envPath := path.Join(path.Dir(path.Dir(filename)), ".env")
	err := godotenv.Load(envPath)
	if err != nil {
		panic("get env file error")
	}
	if len(addrs) == 0 {
		addrs = []string{os.Getenv("ALCHEMY_HTTP_URL"), os.Getenv("ALCHEMY_WS_URL")}
	}
	server := &Server{}
	for _, url := range addrs {
		if strings.HasPrefix(url, "http") {
			server.httpUrl = url
			server.httpClient = &ethClient{url}
		} else if strings.HasPrefix(url, "ws") {
			server.wsUrl = url
			server.wsClient = &ethClient{url}
		} else {
			log.Fatal("Incorrect url format")
		}
	}

	if strings.HasSuffix(server.httpUrl, "8545") {
		log.Println("fetch default account")
		server.accounts = append(server.accounts, core.HexToAddress(FromAddr))
		return server
	}
	return server
}

func (t *Server) HttpClient() *ethClient {
	return t.httpClient
}

func (t *Server) WsClient() *ethClient {
	return t.wsClient
}

// Account returns a specific account
func (t *Server) Account(i int) core.Address {
	return t.accounts[i]
}

// IPCPath returns the ipc endpoint
func (t *Server) IPCPath() string {
	return ""
	// return t.tmpDir + "/geth.ipc"
}

// WSAddr returns the websocket endpoint
func (t *Server) WSAddr() string {
	//return fmt.Sprintf("ws://localhost:8545")
	return t.wsUrl
}

// HTTPAddr returns the http endpoint
func (t *Server) HTTPAddr() string {
	//return fmt.Sprintf(t.addr)
	return t.httpUrl
}

// ProcessBlockWithReceipt ProcessBlock processes a new block
// TODO Finish it
func (t *Server) ProcessBlockWithReceipt() (*core.Receipt, error) {
	return nil, nil
}

// ProcessRawTxWithReceipt ProcessBlock processes a new block via sendrawTransaction
func (t *Server) ProcessRawTxWithReceipt() (*core.Receipt, error) {
	receipt, err := t.SendRawTxn(FromKey, &core.Transaction{
		From:  t.accounts[0],
		To:    &DummyAddr,
		Value: big.NewInt(1e18),
	})
	return receipt, err
}

func (t *Server) ProcessBlock() error {
	_, err := t.ProcessBlockWithReceipt()
	return err
}

func (t *Server) ProcessBlockRaw() error {
	_, err := t.ProcessRawTxWithReceipt()
	return err
}

// Call sends a contract call
func (t *Server) Call(msg *core.CallMsg) (string, error) {
	if isEmptyAddr(msg.From) {
		msg.From = t.Account(0)
	}
	var resp string
	if err := t.httpClient.call("eth_call", &resp, msg, "latest"); err != nil {
		return "", err
	}
	return resp, nil
}

func (t *Server) Fund(address core.Address) (*core.Receipt, error) {
	return t.Transfer(address, big.NewInt(1000000000000000000))
}

// Transfer transfer eth to certain address
// TODO Finish it
func (t *Server) Transfer(address core.Address, value *big.Int) (*core.Receipt, error) {
	return nil, nil
}

// TxnTo sends a transaction to a given method without any arguments
// TODO Finish it
func (t *Server) TxnTo(address core.Address, method string) (*core.Receipt, error) {
	return nil, nil
}

func (t *Server) SendRawTxn(fromKey string, txn *core.Transaction) (*core.Receipt, error) {
	var signer wallet.Signer
	key := wallet.KeyFromString(fromKey)
	if isEmptyAddr(txn.From) {
		txn.From = t.Account(0)
	}
	if txn.GasPrice == 0 {
		txn.GasPrice = DefaultGasPrice
	}
	if txn.Gas == 0 {
		txn.Gas = DefaultGasLimit
	}
	if txn.ChainID != nil {
		signer = wallet.NewEIP155Signer(txn.ChainID.Uint64())
	} else {
		signer = wallet.NewEIP155Signer(core.Local)
	}

	//signer := wallet.NewEIP155Signer(ethgo.Local)
	txn, err := signer.SignTx(txn, key)
	if err != nil {
		return nil, err
	}
	data, err := txn.MarshalRLPTo(nil)
	if err != nil {
		return nil, err
	}
	var hash core.Hash
	hexData := "0x" + hex.EncodeToString(data)
	if err := t.httpClient.call("eth_sendRawTransaction", &hash, hexData); err != nil {
		return nil, err
	}

	return t.WaitForReceipt(hash)
}

// WaitForReceipt waits for the receipt
func (t *Server) WaitForReceipt(hash core.Hash) (*core.Receipt, error) {
	var receipt *core.Receipt
	var count uint64
	// Todo 学习这种loop方法
	for {
		err := t.httpClient.call("eth_getTransactionReceipt", &receipt, hash)
		if err != nil {
			if err.Error() != "not found" {
				return nil, err
			}
		}
		if receipt != nil {
			break
		}
		if count > 300 {
			return nil, fmt.Errorf("timeout waiting for receipt")
		}
		time.Sleep(500 * time.Millisecond)
		count++
	}
	return receipt, nil
}

// DeployContract deploys a contract with account 0 and returns the address
// TODO Finish it
func (t *Server) DeployContract(c *Contract) (*compiler.Artifact, core.Address, error) {
	return nil, core.HexToAddress(""), nil
}
