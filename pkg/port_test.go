package pkg

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetAvailablePort(t *testing.T) {
	port := GetAvailablePort()
	require.Greater(t, port, 1024)
}
