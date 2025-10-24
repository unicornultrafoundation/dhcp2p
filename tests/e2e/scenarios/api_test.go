//go:build e2e

package scenarios

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/unicornultrafoundation/dhcp2p/tests/helpers"
)

func TestAPI_HealthEndpoint(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test")
	}

	// Check if server is already running, otherwise skip the test
	client := helpers.NewHTTPClient("http://localhost:8088")
	resp, err := client.Get("/health")
	if err != nil {
		t.Skipf("Server not available on localhost:8088, skipping health endpoint test: %v", err)
		return
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestAPI_ReadyEndpoint(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test")
	}

	// Check if server is already running, otherwise skip the test
	client := helpers.NewHTTPClient("http://localhost:8088")
	resp, err := client.Get("/ready")
	if err != nil {
		t.Skipf("Server not available on localhost:8088, skipping ready endpoint test: %v", err)
		return
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
