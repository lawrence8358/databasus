package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
)

// The Docker image listens on every interface at 4005; the Windows launcher narrows this to the
// installer's choice (127.0.0.1 and a chosen port by default).
const (
	listenHostEnvVariable = "DATABASUS_LISTEN_HOST"
	portEnvVariable       = "DATABASUS_PORT"
	defaultPort           = 4005
)

type serverAddress struct {
	host string
	port int
}

func getServerAddress() (serverAddress, error) {
	address := serverAddress{host: os.Getenv(listenHostEnvVariable), port: defaultPort}

	if rawPort := os.Getenv(portEnvVariable); rawPort != "" {
		port, err := strconv.Atoi(rawPort)
		if err != nil || port < 1 || port > 65535 {
			return serverAddress{}, fmt.Errorf("%s must be a port number from 1 to 65535, got %q", portEnvVariable, rawPort)
		}

		address.port = port
	}

	return address, nil
}

func (address serverAddress) getListenAddress() string {
	return net.JoinHostPort(address.host, strconv.Itoa(address.port))
}

func (address serverAddress) getLocalBaseURL() string {
	host := address.host

	if ip := net.ParseIP(host); host == "" || ip != nil && ip.IsUnspecified() {
		host = "localhost"
	}

	return "http://" + net.JoinHostPort(host, strconv.Itoa(address.port))
}
