package mock

import (
	"sync"

	"github.com/seth-shi/go-zero-testing-example/app/id/rpc/internal/config"
	"github.com/seth-shi/go-zero-testing-example/app/id/rpc/internal/svc"
)

var SvcCtx = sync.OnceValue(
	func() *svc.ServiceContext {

		return &svc.ServiceContext{
			Config: config.Config{},
			Gen:    IdRpc().Get,
		}
	},
)
