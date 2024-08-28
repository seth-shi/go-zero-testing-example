package mock

import (
	"sync"

	mock2 "github.com/stretchr/testify/mock"
)

var IdRpc = sync.OnceValue(
	func() *mockIdRpc {
		return &mockIdRpc{}
	},
)

type mockIdRpc struct {
	mock2.Mock
}

func (t *mockIdRpc) Get() (uint64, error) {
	args := t.Called()
	return uint64(args.Int(0)), args.Error(1)
}
