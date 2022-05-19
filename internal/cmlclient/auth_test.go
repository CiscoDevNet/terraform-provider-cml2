package cmlclient

import (
	"testing"

	mc "github.com/rschmied/terraform-provider-cml2/m/v2/internal/mockclient"
)

func TestClient_authenticate(t *testing.T) {

	c := NewClient("https://bla.bla", true)
	mclient, ctx := mc.NewMockClient()
	c.httpClient = mclient

	tests := []struct {
		name      string
		userpass  userPass
		responses mc.MockRespList
		wantErr   bool
	}{
		{
			"good",
			userPass{"qwe", "qwe"},
			mc.MockRespList{
				mc.MockResp{
					Data: []byte(`{"version": "2.4.1","ready": true}`),
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
			userPass{"qwe", "qwe"},
			mc.MockRespList{
				mc.MockResp{
					Data: []byte(`"authentication failed"`),
					Code: 403,
				},
			},
			true,
		},
		{
			"noauth",
			userPass{},
			mc.MockRespList{
				mc.MockResp{
					Data: []byte(
						`{
						"description": "Not authenticated: 401 Unauthorized: No authorization token provided.",
						"code": 401					
						}`,
					),
					Code: 401,
				},
			},
			true,
		},
	}
	for _, tt := range tests {
		mclient.SetData(tt.responses)
		t.Run(tt.name, func(t *testing.T) {
			if err := c.authenticate(ctx, tt.userpass); (err != nil) != tt.wantErr {
				t.Errorf("Client.authenticate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
		if !mclient.Empty() {
			t.Error("not all data in mock client consumed")
		}
	}
}
