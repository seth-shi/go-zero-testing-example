package svc

import (
	"github.com/seth-shi/go-zero-testing-example/app/id/rpc/internal/config"
	"github.com/sony/sonyflake"
)

type ServiceContext struct {
	Config config.Config
	Gen    SnowflakeGen
}

type SnowflakeGen func() (uint64, error)

func NewServiceContext(c config.Config) *ServiceContext {

	return &ServiceContext{
		Config: c,
		Gen:    sonyflake.NewSonyflake(sonyflake.Settings{}).NextID,
	}
}
