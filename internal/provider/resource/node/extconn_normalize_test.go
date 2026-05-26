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

	// device-name should be normalized to label and marked changed
	norm, changed, warn, err := normalizeExtConnConfig(context.Background(), cfg, "virbr0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !changed || norm != "NAT" || warn == "" {
		t.Fatalf("normalize mismatch: got=%q changed=%v warn=%q", norm, changed, warn)
	}

	// exact label should not be changed
	norm2, changed2, _, err := normalizeExtConnConfig(context.Background(), cfg, "System Bridge")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if changed2 || norm2 != "System Bridge" {
		t.Fatalf("expected no-change for label: got=%q changed=%v", norm2, changed2)
	}

	// empty / whitespace input is a no-op — List is never called
	ne, nc, nw, err := normalizeExtConnConfig(context.Background(), cfg, "   ")
	if err != nil {
		t.Fatalf("unexpected error for empty input: %v", err)
	}
	if ne != "" || nc || nw != "" {
		t.Fatalf("expected empty no-op for whitespace: got=%q changed=%v warn=%q", ne, nc, nw)
	}

	// unknown connector — no device-name or label match, returned as-is (no change)
	nu, ncu, _, err := normalizeExtConnConfig(context.Background(), cfg, "unknown-conn")
	if err != nil {
		t.Fatalf("unexpected error for unknown input: %v", err)
	}
	if ncu || nu != "unknown-conn" {
		t.Fatalf("expected no-change for unknown connector: got=%q changed=%v", nu, ncu)
	}

	// API error path: server returns 500 — error must propagate
	mode = 1
	if _, _, _, err := normalizeExtConnConfig(context.Background(), cfg, "virbr0"); err == nil {
		t.Fatalf("expected error when API returns 500")
	}
}

func TestNormalizeExtConnConfig_NilGuards(t *testing.T) {
	ctx := context.Background()

	// nil cfg — concrete pointer nil, check is reliable
	_, _, _, err := normalizeExtConnConfig(ctx, nil, "virbr0")
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
	_, _, _, err = normalizeExtConnConfig(ctx, uninit, "virbr0")
	if err == nil {
		t.Fatal("expected error for uninitialized cfg (nil client)")
	}
}
