package svc

import (
	"testing"

	"github.com/seth-shi/go-zero-testing-example/app/id/rpc/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestNewServiceContext(t *testing.T) {

	svcCtx := NewServiceContext(config.Config{})
	assert.NotNil(t, svcCtx)
}
