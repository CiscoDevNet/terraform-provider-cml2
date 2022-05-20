package cmlclient

import (
	"testing"

	mc "github.com/rschmied/terraform-provider-cml2/m/v2/internal/mockclient"
)

func TestClient_authenticate(t *testing.T) {

	c := NewClient("https://bla.bla", true)
	mclient, ctx := mc.NewMockClient()
	c.httpClient = mclient
	c.authChecked = false
	c.versionChecked = true
	c.userpass = userPass{"qwe", "qwe"}

	tests := []struct {
		name      string
		responses mc.MockRespList
		wantErr   bool
	}{
		{
			"good",
			mc.MockRespList{
				mc.MockResp{
					Data: []byte(`{
						"description": "Not authenticated: 401 Unauthorized: No authorization token provided.",
						"code":        401
					}`),
					Code: 401,
				},
				mc.MockResp{
					Data: []byte(`{"username": "qwe", "id": "008", "token": "secret" }`),
					Code: 200,
				},
				mc.MockResp{
					Data: []byte(`"OK"`),
					Code: 200,
				},
				mc.MockResp{
					Data: []byte(`{"username": "qwe", "id": "008", "token": "secret" }`),
					Code: 200,
				},
			},
			false,
		},
		{
			"bad",
			mc.MockRespList{
				mc.MockResp{
					Data: []byte(`{
						"description": "Not authenticated: 401 Unauthorized: No authorization token provided.",
						"code":        401
					}`),
					Code: 401,
				},
				mc.MockResp{
					Data: []byte(`"authentication failed"`),
					Code: 403,
				},
			},
			true,
		},
	}
	for _, tt := range tests {
		mclient.SetData(tt.responses)
		t.Run(tt.name, func(t *testing.T) {
			if err := c.authenticate(ctx, c.userpass); (err != nil) != tt.wantErr {
				t.Errorf("Client.authenticate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
		if !mclient.Empty() {
			t.Error("not all data in mock client consumed")
		}
	}
}
