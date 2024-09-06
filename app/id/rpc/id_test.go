package main

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/samber/lo"
	"github.com/seth-shi/go-zero-testing-example/app/id/rpc/id"
	"github.com/seth-shi/go-zero-testing-example/pkg"
	"github.com/stretchr/testify/require"
	"github.com/zeromicro/go-zero/zrpc"
)

func Test_get(t *testing.T) {
	var (
		ctx = context.Background()
	)
	client := newRpcClient(t)
	resp, err := client.Get(ctx, &id.IdRequest{})
	require.NoError(t, err)
	resp2, err2 := client.Get(ctx, &id.IdRequest{})
	require.NoError(t, err2)
	require.Greater(t, resp2.GetId(), resp.GetId())
}

var (
	testListenOn string
)

func TestMain(m *testing.M) {

	testListenOn = fmt.Sprintf(":%d", pkg.GetAvailablePort())
	data := fmt.Sprintf(
		`
Name: id.rpc
ListenOn: %s
`, testListenOn,
	)
	remove, tmp := lo.Must2(pkg.CreateTempFile(".yaml", data))
	defer remove()
	configFile = tmp

	go main()

	os.Exit(m.Run())
}

func newRpcClient(t *testing.T) id.IdClient {
	t.Helper()

	conn, err := zrpc.NewClient(
		zrpc.RpcClientConf{
			Target:   testListenOn,
			NonBlock: false,
		},
	)
	require.NoError(t, err)
	return id.NewIdClient(conn.Conn())
}
