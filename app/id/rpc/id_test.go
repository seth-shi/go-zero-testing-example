package main

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/seth-shi/go-zero-testing-example/app/id/rpc/id"
	"github.com/seth-shi/go-zero-testing-example/app/id/rpc/internal/config"
	"github.com/seth-shi/go-zero-testing-example/pkg"
	"github.com/stretchr/testify/require"
	"github.com/zeromicro/go-zero/zrpc"
)

func Test_loadConfigByFile(t *testing.T) {
	remove, tmp, err := pkg.CreateTempFile(
		".yaml", `
Name: id.rpc
ListenOn: 0.0.0.0:9502
`,
	)
	require.NoError(t, err)
	defer remove()

	svcCtx := loadConfigByFile(tmp)
	require.NoError(t, err)
	require.Equal(t, "id.rpc", svcCtx.Name)
}

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

	getConfig = func(filename string) config.Config {
		return config.Config{
			RpcServerConf: zrpc.RpcServerConf{
				ListenOn: testListenOn,
			},
		}
	}

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
