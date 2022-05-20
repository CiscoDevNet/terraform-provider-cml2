package cmlclient

import (
	"testing"

	mc "github.com/rschmied/terraform-provider-cml2/m/v2/internal/mockclient"
)

func TestClient_VersionCheck(t *testing.T) {

	c := NewClient("https://bla.bla", true)
	mclient, ctx := mc.NewMockClient()
	c.httpClient = mclient
	c.versionChecked = true
	c.authChecked = true

	tests := []struct {
		name       string
		wantJSON   string
		wantStatus int
		wantErr    bool
	}{
		{
			"too old", `{"version": "2.1.0","ready": true}`, 200, true,
		},
		{
			"garbage", `{"version": "garbage","ready": true}`, 200, true,
		},
		{
			"too new", `{"version": "2.35.0","ready": true}`, 200, true,
		},
		{
			"perfect", `{"version": "2.4.0","ready": true}`, 200, false,
		},
		{
			"newer", `{"version": "2.4.1","ready": true}`, 200, false,
		},
		{
			"devbuild", `{"version": "2.4.0.dev0+build.f904bdf8","ready": true}`, 200, false,
		},
	}
	for _, tt := range tests {
		mclient.SetData(mc.MockRespList{
			{Data: []byte(tt.wantJSON), Code: tt.wantStatus},
		})
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
