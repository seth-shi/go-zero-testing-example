package main

import (
	"context"
	"os"
	"testing"

	"github.com/samber/lo"
	"github.com/seth-shi/go-zero-testing-example/app/post/rpc/internal/config"
	"github.com/seth-shi/go-zero-testing-example/app/post/rpc/internal/faker"
	"github.com/seth-shi/go-zero-testing-example/app/post/rpc/internal/svc"
	"github.com/seth-shi/go-zero-testing-example/app/post/rpc/post"
	"github.com/seth-shi/go-zero-testing-example/pkg"
	"github.com/stretchr/testify/require"
	"github.com/zeromicro/go-zero/zrpc"
)

func TestMain(m *testing.M) {

	// 使用默认配置
	svcCtxGet = func(c config.Config) *svc.ServiceContext {
		return faker.GetValue().SvcCtx
	}

	data := `Name: post.rpc
ListenOn: 0.0.0.0:8080
DataSource: 127.0.0.1:3306?charset=utf8mb4&parseTime=true&loc=Local
IdRpc:
  Target: 0.0.0.0:8081
RedisConf:
  Host: 127.0.0.1:6379
`
	remove, tmp := lo.Must2(pkg.CreateTempFile(".yaml", data))
	defer remove()
	configFile = tmp

	go main()

	// 运行测试
	code := m.Run()
	os.Exit(code)
}

// 集成测试
func TestGet(t *testing.T) {

	var (
		fakerVal  = faker.GetValue()
		postModel = fakerVal.Models.PostModel
	)
	conn, err := zrpc.NewClient(
		zrpc.RpcClientConf{
			Target:   fakerVal.RpcListen,
			NonBlock: false,
		},
	)
	require.NoError(t, err)
	client := post.NewPostClient(conn.Conn())
	resp, err := client.Get(context.Background(), &post.PostRequest{Id: postModel.ID})
	require.NoError(t, err)
	require.NotZero(t, resp.GetId())
	require.Equal(t, resp.GetId(), postModel.ID)
	require.Equal(t, resp.Title, lo.FromPtr(postModel.Title))
}
