package pkg

import (
	"errors"
	"net"
)

var (
	errNotAddress = errors.New("listener address is not tcp address")
)

func GetAvailablePort() (int, error) {
	address, err := net.ResolveTCPAddr("tcp", "0.0.0.0:0")
	if err != nil {
		return 0, err
	}

	listener, err := net.ListenTCP("tcp", address)
	if err != nil {
		return 0, err
	}
	//nolint:errcheck
	defer listener.Close()

	addr, ok := listener.Addr().(*net.TCPAddr)
	if !ok {
		return 0, errNotAddress
	}

	return addr.Port, nil
}
