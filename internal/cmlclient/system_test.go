package cmlclient

import (
	"testing"

	mc "github.com/rschmied/terraform-provider-cml2/m/v2/internal/mockresponder"
)

func TestClient_VersionCheck(t *testing.T) {

	c := NewClient("https://bla.bla", true)
	mclient, ctx := mc.NewMockResponder()
	c.httpClient = mclient
	c.versionChecked = true
	c.authChecked = true

	tests := []struct {
		name     string
		wantJSON string
		wantErr  bool
	}{
		{"too old", `{"version": "2.1.0","ready": true}`, true},
		{"garbage", `{"version": "garbage","ready": true}`, true},
		{"too new", `{"version": "2.35.0","ready": true}`, true},
		{"perfect", `{"version": "2.4.0","ready": true}`, false},
		{"actual", `{"version": "2.4.0+build.1","ready": true}`, false},
		{"newer", `{"version": "2.4.1","ready": true}`, false},
		{"devbuild", `{"version": "2.4.0.dev0","ready": true}`, false},
	}
	for _, tt := range tests {
		mclient.SetData(mc.MockRespList{{Data: []byte(tt.wantJSON)}})
		t.Run(tt.name, func(t *testing.T) {
			if err := c.versionCheck(ctx); (err != nil) != tt.wantErr {
				t.Errorf("Client.VersionCheck() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
		if !mclient.Empty() {
			t.Error("not all data in mock client consumed")
		}
	}
}
