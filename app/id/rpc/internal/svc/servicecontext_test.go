package svc

import (
	"testing"

	"github.com/seth-shi/go-zero-testing-example/app/id/rpc/internal/config"
	"github.com/stretchr/testify/require"
)

func TestNewServiceContext(t *testing.T) {

	svcCtx := NewServiceContext(config.Config{})
	require.NotNil(t, svcCtx)
}
