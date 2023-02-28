package testutil

import (
	"fmt"
	"github.com/deep-nl/ethgo"
	"testing"
)

//func TestDeployServer(t *testing.T) {
//	srv := DeployTestServer(t, nil)
//	require.NotEmpty(t, srv.accounts)
//
//	clt := &ethClient{srv.HTTPAddr()}
//	account := []ethgo.Address{}
//
//	err := clt.call("eth_accounts", &account)
//	require.NoError(t, err)
//}

func TestTestServer_Account(t *testing.T) {
	server := &TestServer{}
	tx := &ethgo.Transaction{}

	fmt.Println(server.HTTPAddr() == "")
	fmt.Println(server.WSAddr())
	fmt.Printf("%v", tx.ChainID == nil)

}
