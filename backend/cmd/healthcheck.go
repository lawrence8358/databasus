package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const healthcheckTimeout = 4 * time.Second

func checkServerAlive(baseURL string) error {
	ctx, cancel := context.WithTimeout(context.Background(), healthcheckTimeout)
	defer cancel()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/api/v1/system/version", nil)
	if err != nil {
		return err
	}

	response, err := newHealthcheckClient(baseURL).Do(request)
	if err != nil {
		return err
	}
	defer func() { _ = response.Body.Close() }()

	_, _ = io.Copy(io.Discard, response.Body)

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", response.StatusCode)
	}

	return nil
}

// The healthcheck asks this computer's own server whether it is up; it trusts nothing it receives,
// so it accepts the self-signed certificate an installer may have generated.
func newHealthcheckClient(baseURL string) *http.Client {
	if !strings.HasPrefix(baseURL, "https://") {
		return http.DefaultClient
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec // liveness probe of the local server

	return &http.Client{Transport: transport}
}

func runHealthcheckCommand() {
	address, err := getServerAddress()
	if err == nil {
		err = checkServerAlive(address.getLocalBaseURL())
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "healthcheck failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("healthcheck ok")
	os.Exit(0)
}
