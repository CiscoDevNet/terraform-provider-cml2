package node

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rschmied/gocmlclient/pkg/models"

	"github.com/ciscodevnet/terraform-provider-cml2/internal/cmlschema"
	"github.com/ciscodevnet/terraform-provider-cml2/internal/common"
)

// newTestProviderConfig spins up a TLS httptest server using the provided mux
// and returns an initialized *common.ProviderConfig that points at it.
// The caller is responsible for calling srv.Close().
func newTestProviderConfig(t *testing.T, mux *http.ServeMux) (*common.ProviderConfig, *httptest.Server) {
	t.Helper()
	srv := httptest.NewTLSServer(mux)

	model := &cmlschema.ProviderModel{
		Address:        types.StringValue(srv.URL),
		Token:          types.StringValue("test-token"),
		SkipVerify:     types.BoolValue(true),
		NamedConfigs:   types.BoolValue(false),
		UseCache:       types.BoolNull(),
		TokenCache:     types.BoolValue(false),
		TokenCacheFile: types.StringNull(),
		CAcert:         types.StringNull(),
		Username:       types.StringNull(),
		Password:       types.StringNull(),
		RequestHeaders: types.MapNull(types.StringType),
		DynamicConfig:  types.BoolNull(),
	}

	var diags diag.Diagnostics
	cfg := common.NewProviderConfig(model).Initialize(context.Background(), &diags)
	if diags.HasError() {
		srv.Close()
		t.Fatalf("provider config initialization failed: %v", diags.Errors())
	}
	return cfg, srv
}

func extConnMux(t *testing.T, conns []*models.ExtConn, mode *int) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v0/system/external_connectors", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("unexpected method: %s", r.Method)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if *mode == 1 {
			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(conns)
	})
	return mux
}

func TestNormalizeExtConnConfig(t *testing.T) {
	conns := []*models.ExtConn{
		{ID: models.UUID("1"), DeviceName: "virbr0", Label: "NAT"},
		{ID: models.UUID("2"), DeviceName: "bridge0", Label: "System Bridge"},
	}

	mode := 0
	cfg, srv := newTestProviderConfig(t, extConnMux(t, conns, &mode))
	defer srv.Close()

	// Device name is accepted as-is.
	norm, err := normalizeExtConnConfig(context.Background(), cfg, "virbr0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if norm != "virbr0" {
		t.Fatalf("expected device name passthrough, got=%q", norm)
	}

	// Label must be rejected with guidance.
	if _, err := normalizeExtConnConfig(context.Background(), cfg, "System Bridge"); err == nil {
		t.Fatalf("expected error for label input")
	}

	// Empty input is a no-op.
	nullv, err := normalizeExtConnConfig(context.Background(), cfg, "   ")
	if err != nil {
		t.Fatalf("unexpected error for empty input: %v", err)
	}
	if nullv != "" {
		t.Fatalf("expected empty no-op for whitespace: got=%q", nullv)
	}

	// Unknown connector should error.
	if _, err := normalizeExtConnConfig(context.Background(), cfg, "unknown-conn"); err == nil {
		t.Fatalf("expected error for unknown connector")
	}

	// API error path: server returns 500 — error must propagate.
	mode = 1
	if _, err := normalizeExtConnConfig(context.Background(), cfg, "virbr0"); err == nil {
		t.Fatalf("expected error when API returns 500")
	}
}

func TestNormalizeExtConnConfig_NilGuards(t *testing.T) {
	ctx := context.Background()

	// nil cfg — concrete pointer nil, check is reliable
	_, err := normalizeExtConnConfig(ctx, nil, "virbr0")
	if err == nil {
		t.Fatal("expected error for nil cfg")
	}

	// uninitialised ProviderConfig — cfg.Client() returns nil
	uninit := common.NewProviderConfig(&cmlschema.ProviderModel{
		Address:        types.StringValue("https://localhost"),
		Token:          types.StringValue("t"),
		SkipVerify:     types.BoolValue(true),
		NamedConfigs:   types.BoolValue(false),
		UseCache:       types.BoolNull(),
		TokenCache:     types.BoolValue(false),
		TokenCacheFile: types.StringNull(),
		CAcert:         types.StringNull(),
		Username:       types.StringNull(),
		Password:       types.StringNull(),
		RequestHeaders: types.MapNull(types.StringType),
		DynamicConfig:  types.BoolNull(),
	})
	_, err = normalizeExtConnConfig(ctx, uninit, "virbr0")
	if err == nil {
		t.Fatal("expected error for uninitialized cfg (nil client)")
	}
}
