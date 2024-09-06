package faker

import (
	"fmt"
	"sync"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/samber/lo"
	"github.com/seth-shi/go-zero-testing-example/app/id/rpc/id"
	"github.com/seth-shi/go-zero-testing-example/app/post/rpc/internal/config"
	"github.com/seth-shi/go-zero-testing-example/app/post/rpc/internal/model/do"
	"github.com/seth-shi/go-zero-testing-example/app/post/rpc/internal/svc"
	"github.com/seth-shi/go-zero-testing-example/pkg"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const (
	startId = 100000
)

type value struct {
	IdServer    *idServer
	Redis       *miniredis.Miniredis
	Models      *fakerModels
	Gorm        *gorm.DB
	DatabaseDsn string
	RedisAddr   string
	RpcListen   string
	SvcCtx      *svc.ServiceContext
}

var GetValue = sync.OnceValue(
	func() value {

		rpcPort := pkg.GetAvailablePort()
		redisMock, redisAddr := pkg.FakerRedisServer()
		mysqlDsn := pkg.FakerDatabaseServer()
		dbConn := lo.Must(gorm.Open(mysql.Open(mysqlDsn)))

		testIdServer := newIdServer()
		listenOn := fmt.Sprintf(":%d", rpcPort)

		return value{
			IdServer:    testIdServer,
			Redis:       redisMock,
			RedisAddr:   redisAddr,
			Models:      makeDatabase(dbConn),
			RpcListen:   listenOn,
			DatabaseDsn: mysqlDsn,
			Gorm:        dbConn,
			SvcCtx: &svc.ServiceContext{
				Config: config.Config{
					RpcServerConf: zrpc.RpcServerConf{
						ListenOn: listenOn,
						ServiceConf: service.ServiceConf{
							Mode: service.TestMode,
						},
					},
				},
				Redis: redis.NewClient(&redis.Options{Addr: redisAddr}),
				IdRpc: id.NewIdClient(testIdServer.Connect()),
				Query: do.Use(dbConn),
			},
		}
	},
)
