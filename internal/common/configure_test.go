package common_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ciscodevnet/terraform-provider-cml2/internal/cmlschema"
	"github.com/ciscodevnet/terraform-provider-cml2/internal/common"
)

func TestProviderConfigInitialize_RequestHeaders(t *testing.T) {
	t.Parallel()

	var authHeader string
	var systemHeader string

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v0/auth_extended":
			authHeader = r.Header.Get("X-Proxy-Token")
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprint(w, `{"id":"1","username":"user","token":"jwt-token","admin":false}`)
		case "/api/v0/system_information":
			systemHeader = r.Header.Get("X-Proxy-Token")
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprint(w, `{"version":"2.9.0","ready":true}`)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	headers := types.MapValueMust(types.StringType, map[string]attr.Value{
		"X-Proxy-Token": types.StringValue("proxy-secret"),
	})

	data := &cmlschema.ProviderModel{
		Address:        types.StringValue(server.URL),
		Username:       types.StringValue("user"),
		Password:       types.StringValue("secret"),
		RequestHeaders: headers,
		SkipVerify:     types.BoolValue(true),
		NamedConfigs:   types.BoolValue(false),
		TokenCache:     types.BoolValue(false),
		UseCache:       types.BoolValue(false),
	}

	config := common.NewProviderConfig(data)
	var diags diag.Diagnostics
	config.Initialize(context.Background(), &diags)
	require.False(t, diags.HasError(), diags.Errors())
	require.NotNil(t, config.Client())

	err := config.Client().System.Ready(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "proxy-secret", authHeader)
	assert.Equal(t, "proxy-secret", systemHeader)
}
