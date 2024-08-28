package svc

import (
	"github.com/redis/go-redis/v9"
	"github.com/seth-shi/go-zero-testing-example/app/id/rpc/id"
	"github.com/seth-shi/go-zero-testing-example/app/post/rpc/internal/config"
	"github.com/seth-shi/go-zero-testing-example/app/post/rpc/internal/model/do"
	"github.com/seth-shi/go-zero-testing-example/app/post/rpc/internal/model/entity"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config config.Config
	Redis  *redis.Client
	IdRpc  id.IdClient

	Query *do.Query
}

func NewServiceContext(c config.Config) *ServiceContext {

	conn, err := gorm.Open(mysql.Open(c.DataSource))
	logx.Must(err)

	idClient := id.NewIdClient(zrpc.MustNewClient(c.IdRpc).Conn())
	entity.SetIdGenerator(idClient)

	rdb := redis.NewClient(
		&redis.Options{
			Addr:     c.RedisConf.Host,
			Password: c.RedisConf.Pass,
			DB:       0,
		},
	)

	return &ServiceContext{
		Config: c,
		Redis:  rdb,
		IdRpc:  idClient,
		Query:  do.Use(conn),
	}
}
