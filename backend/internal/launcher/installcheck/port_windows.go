//go:build windows

package installcheck

import (
	"fmt"
	"net"
	"strconv"
)

// CheckPortIsFree returns nil when nothing listens on the port, either on the loopback address or
// on every interface, the two places the Web UI can listen.
func CheckPortIsFree(port int) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("the port must be from 1 to 65535, got %d", port)
	}

	for _, host := range []string{"127.0.0.1", ""} {
		listener, err := net.Listen("tcp", net.JoinHostPort(host, strconv.Itoa(port)))
		if err != nil {
			return fmt.Errorf("port %d is already in use by another program", port)
		}

		if err := listener.Close(); err != nil {
			return err
		}
	}

	return nil
}
