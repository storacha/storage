package testutil

import (
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

// GetFreePort asks the kernel for a free open port that is ready to use.
func GetFreePort(t *testing.T) int {
	a, err := net.ResolveTCPAddr("tcp", "localhost:0")
	require.NoError(t, err)
	l, err := net.ListenTCP("tcp", a)
	require.NoError(t, err)
	port := l.Addr().(*net.TCPAddr).Port
	err = l.Close()
	require.NoError(t, err)
	return port
}
