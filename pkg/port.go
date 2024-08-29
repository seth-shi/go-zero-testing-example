package pkg

import (
	"errors"
	"net"

	"github.com/zeromicro/go-zero/core/logx"
)

var (
	errNotAddress = errors.New("listener address is not tcp address")
)

func GetAvailablePort() int {
	address, err := net.ResolveTCPAddr("tcp", "0.0.0.0:0")
	logx.Must(err)

	listener, err := net.ListenTCP("tcp", address)
	logx.Must(err)

	//nolint:errcheck
	defer listener.Close()

	addr, ok := listener.Addr().(*net.TCPAddr)
	if !ok {
		logx.Must(errNotAddress)
	}

	return addr.Port
}
