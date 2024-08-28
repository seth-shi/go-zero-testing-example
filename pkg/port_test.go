package pkg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAvailablePort(t *testing.T) {
	port, err := GetAvailablePort()
	assert.NoError(t, err)
	assert.Greater(t, port, 1024)
}
