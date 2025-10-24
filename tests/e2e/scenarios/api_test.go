//go:build e2e

package scenarios

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unicornultrafoundation/dhcp2p/tests/helpers"
)

func TestAPI_HealthEndpoint(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test")
	}

	client := helpers.NewHTTPClient("http://localhost:8088")

	resp, err := client.Get("/health")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestAPI_ReadyEndpoint(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test")
	}

	client := helpers.NewHTTPClient("http://localhost:8088")

	resp, err := client.Get("/ready")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
