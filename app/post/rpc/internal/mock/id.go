package mock

import (
	"context"

	"github.com/seth-shi/go-zero-testing-example/app/id/rpc/id"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

type idMock struct {
	mock.Mock
}

func (m *idMock) Get(ctx context.Context, in *id.IdRequest, opts ...grpc.CallOption) (*id.IdResponse, error) {
	args := m.Called()
	idResp := uint64(args.Int(0))

	return &id.IdResponse{
		Id:   idResp,
		Node: idResp,
	}, args.Error(1)
}
