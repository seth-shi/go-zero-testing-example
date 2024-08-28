package main

import (
	"context"
	"os"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/samber/lo"
	"github.com/seth-shi/go-zero-testing-example/app/post/rpc/internal/config"
	"github.com/seth-shi/go-zero-testing-example/app/post/rpc/internal/faker"
	"github.com/seth-shi/go-zero-testing-example/app/post/rpc/internal/model/do"
	"github.com/seth-shi/go-zero-testing-example/app/post/rpc/internal/svc"
	"github.com/seth-shi/go-zero-testing-example/app/post/rpc/post"
	"github.com/stretchr/testify/assert"
	"github.com/zeromicro/go-zero/zrpc"
)

func TestMain(m *testing.M) {

	// 使用默认配置
	svcCtxGet = func() (*svc.ServiceContext, error) {

		fakerVal := faker.GetValue()
		return &svc.ServiceContext{
			Config: config.Config{
				RpcServerConf: zrpc.RpcServerConf{
					ListenOn: fakerVal.RpcListen,
				},
			},
			Redis: redis.NewClient(
				&redis.Options{
					Addr: fakerVal.RedisAddr,
					DB:   0,
				},
			),
			IdRpc: fakerVal.IdServer,
			Query: do.Use(fakerVal.Gorm),
		}, nil
	}

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
	assert.NoError(t, err)
	client := post.NewPostClient(conn.Conn())
	resp, err := client.Get(context.Background(), &post.PostRequest{Id: postModel.ID})
	assert.NoError(t, err)
	assert.NotZero(t, resp.GetId())
	assert.Equal(t, resp.GetId(), postModel.ID)
	assert.Equal(t, resp.Title, lo.FromPtr(postModel.Title))
}
