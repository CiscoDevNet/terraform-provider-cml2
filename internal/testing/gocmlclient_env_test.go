package testing

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCMLClientFromTFEnvTrimsToken(t *testing.T) {
	var authorization string

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v0/system_information" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		authorization = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"version":"2.11.0","ready":true}`)
	}))
	defer server.Close()

	t.Setenv("TF_VAR_address", server.URL)
	t.Setenv("TF_VAR_token", "jwt-token\n")
	t.Setenv("TF_VAR_username", "")
	t.Setenv("TF_VAR_password", "")

	client, err := NewCMLClientFromTFEnv()
	require.NoError(t, err)
	require.NoError(t, client.System.Ready(context.Background()))
	assert.Equal(t, "Bearer jwt-token", authorization)
}
