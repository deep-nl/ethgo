package jsonrpc

import (
	"github.com/deep-nl/ethgo/core"
	"strings"
	"testing"
	"time"

	"github.com/deep-nl/ethgo/testutil"
	"github.com/stretchr/testify/assert"
)

func TestSubscribeNewHead(t *testing.T) {
	testutil.MultiAddr(t, func(s *testutil.TestServer, addr string) {
		if strings.HasPrefix(addr, "http") {
			t.Log("wrong url")
			return
		}

		c, _ := NewClient(addr)
		//c.SetMaxConnsLimit(0)
		defer c.Close()

		data := make(chan []byte)
		cancel, err := c.Subscribe("newHeads", func(b []byte) {
			data <- b
		})
		assert.NoError(t, err)

		var lastBlock *core.Block
		recv := func(ok bool) {
			select {
			case buf := <-data:
				if !ok {
					t.Fatal("unexpected value")
				}

				var block core.Block
				if err := block.UnmarshalJSON(buf); err != nil {
					t.Fatal(err)
				}
				if lastBlock != nil {
					if lastBlock.Number+1 != block.Number {
						t.Fatalf("bad sequence %d %d", lastBlock.Number, block.Number)
					}
				}
				lastBlock = &block
				t.Logf("Blocknumbe %v", lastBlock.Number)

			case <-time.After(1 * time.Second):
				if ok {
					t.Fatal("timeout for new head")
				}
			}
		}

		// 重新把addr改为http
		s = testutil.NewTestServer(t, "http://127.0.0.1:8545")

		err = s.ProcessBlockRaw()
		assert.NoError(t, err)

		recv(true)

		err = s.ProcessBlockRaw()
		assert.NoError(t, err)
		recv(true)
		t.Logf("Blocknumbe %v", lastBlock.Number)

		assert.NoError(t, cancel())

		err = s.ProcessBlockRaw()
		assert.NoError(t, err)
		recv(false)

		// subscription already closed
		assert.Error(t, cancel())
	})
}
