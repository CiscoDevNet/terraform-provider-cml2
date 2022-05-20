package cmlclient

import (
	"testing"

	mr "github.com/rschmied/terraform-provider-cml2/m/v2/internal/mockresponder"
)

func TestClient_authenticate(t *testing.T) {

	c := NewClient("https://bla.bla", true)
	mresp, ctx := mr.NewMockResponder()
	c.httpClient = mresp
	c.authChecked = false
	c.versionChecked = true
	c.userpass = userPass{"qwe", "qwe"}

	tests := []struct {
		name      string
		responses mr.MockRespList
		wantErr   bool
	}{
		{
			"good",
			mr.MockRespList{
				mr.MockResp{
					Data: []byte(`{
						"description": "Not authenticated: 401 Unauthorized: No authorization token provided.",
						"code":        401
					}`),
					Code: 401,
				},
				mr.MockResp{
					Data: []byte(`{"username": "qwe", "id": "008", "token": "secret" }`),
				},
				mr.MockResp{
					Data: []byte(`"OK"`),
				},
				mr.MockResp{
					Data: []byte(`{"username": "qwe", "id": "008", "token": "secret" }`),
				},
			},
			false,
		},
		{
			"bad",
			mr.MockRespList{
				mr.MockResp{
					Data: []byte(`{
						"description": "Not authenticated: 401 Unauthorized: No authorization token provided.",
						"code":        401
					}`),
					Code: 401,
				},
				mr.MockResp{
					Data: []byte(`"authentication failed"`),
					Code: 403,
				},
			},
			true,
		},
	}
	for _, tt := range tests {
		mresp.SetData(tt.responses)
		t.Run(tt.name, func(t *testing.T) {
			if err := c.authenticate(ctx, c.userpass); (err != nil) != tt.wantErr {
				t.Errorf("Client.authenticate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
		if !mresp.Empty() {
			t.Error("not all data in mock client consumed")
		}
	}
}
