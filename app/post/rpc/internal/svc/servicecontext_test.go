package svc

import (
	"testing"

	"github.com/seth-shi/go-zero-testing-example/app/post/rpc/internal/config"
	"github.com/seth-shi/go-zero-testing-example/pkg"
	"github.com/stretchr/testify/require"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

func TestNewServiceContext(t *testing.T) {
	var (
		dsn          = pkg.FakerDatabaseServer()
		_, redisAddr = pkg.FakerRedisServer()
		c            = config.Config{
			RpcServerConf: zrpc.RpcServerConf{},
			DataSource:    dsn,
			RedisConf: redis.RedisConf{
				Host: redisAddr,
			},
			IdRpc: zrpc.RpcClientConf{
				Target:   "127.0.0.1:3000",
				NonBlock: true,
			},
		}
	)
	ctx := NewServiceContext(c)
	require.NotNil(t, ctx)
}
