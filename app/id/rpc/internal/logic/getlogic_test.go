package logic

import (
	"context"
	"errors"
	"testing"

	"github.com/seth-shi/go-zero-testing-example/app/id/rpc/id"
	"github.com/seth-shi/go-zero-testing-example/app/id/rpc/internal/mock"
	mock2 "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewGetLogic(t *testing.T) {

	logic := NewGetLogic(context.Background(), mock.SvcCtx())
	mockRpc := mock.IdRpc()
	mockCall := mockRpc.On("Get", mock2.Anything).Return(1, nil)
	resp, err := logic.Get(&id.IdRequest{})
	require.NoError(t, err)
	require.NotZero(t, resp.GetId())

	// mock 错误
	mockCall.Unset()
	mockRpc.On("Get", mock2.Anything).Return(0, errors.New("wrong"))
	_, err3 := logic.Get(&id.IdRequest{})
	require.Error(t, err3)
}
