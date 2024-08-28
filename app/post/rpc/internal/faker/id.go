package faker

import (
	"context"
	"sync"

	"github.com/seth-shi/go-zero-testing-example/app/id/rpc/id"
	"google.golang.org/grpc"
)

type idGenerator struct {
	startId uint64
	locker  sync.Locker
}

func (m *idGenerator) Get(ctx context.Context, in *id.IdRequest, opts ...grpc.CallOption) (*id.IdResponse, error) {

	m.locker.Lock()
	defer m.locker.Unlock()

	m.startId++

	return &id.IdResponse{
		Id:   m.startId,
		Node: 1,
	}, nil
}
