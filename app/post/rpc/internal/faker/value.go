package faker

import (
	"fmt"
	"math/rand"
	"sync"

	"github.com/alicebob/miniredis/v2"
	"github.com/seth-shi/go-zero-testing-example/pkg"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type value struct {
	IdServer    *idGenerator
	Redis       *miniredis.Miniredis
	Models      *fakerModels
	Gorm        *gorm.DB
	RedisAddr   string
	RpcListen   string
	DatabaseDsn string
}

var GetValue = sync.OnceValue(
	func() value {

		redis, redisAddr := pkg.FakerRedisServer()
		dsn := pkg.FakerDatabaseServer()

		rpcPort := pkg.GetAvailablePort()

		conn, err := gorm.Open(mysql.Open(dsn))
		logx.Must(err)

		idGen := &idGenerator{
			startId: uint64(rand.Int() + 1),
			locker:  &sync.RWMutex{},
		}
		return value{
			IdServer:    idGen,
			Redis:       redis,
			RedisAddr:   redisAddr,
			DatabaseDsn: dsn,
			Models:      makeDatabase(dsn, idGen),
			RpcListen:   fmt.Sprintf(":%d", rpcPort),
			Gorm:        conn,
		}
	},
)
