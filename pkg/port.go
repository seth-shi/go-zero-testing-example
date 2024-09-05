package pkg

import (
	"errors"
	"net"

	"github.com/samber/lo"
)

var (
	errNotAddress = errors.New("listener address is not tcp address")
)

func GetAvailablePort() int {
	address, err := net.ResolveTCPAddr("tcp", "0.0.0.0:0")
	lo.Must0(err)

	listener, err := net.ListenTCP("tcp", address)
	lo.Must0(err)

	//nolint:errcheck
	defer listener.Close()

	addr, ok := listener.Addr().(*net.TCPAddr)
	lo.Must0(lo.Validate(ok, errNotAddress.Error()))

	return addr.Port
}
