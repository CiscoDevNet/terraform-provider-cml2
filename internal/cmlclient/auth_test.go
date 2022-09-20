package cmlclient

import (
	"errors"
	"testing"

	mr "github.com/rschmied/terraform-provider-cml2/m/v2/internal/mockresponder"
	"github.com/stretchr/testify/assert"
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

func TestClient_token_auth(t *testing.T) {

	c := NewClient("https://bla.bla", true)
	mresp, ctx := mr.NewMockResponder()
	c.httpClient = mresp
	c.authChecked = false
	c.versionChecked = true
	c.apiToken = "sometoken"

	tests := []struct {
		name      string
		responses mr.MockRespList
		wantErr   bool
		errstr    string
	}{
		{
			"goodtoken",
			mr.MockRespList{
				mr.MockResp{
					Data: []byte(`"OK"`),
					Code: 200,
				},
				mr.MockResp{
					Data: []byte(`{"version": "2.4.1","ready": true}`),
					Code: 200,
				},
			},
			false,
			"",
		},
		{
			"badjson",
			mr.MockRespList{
				mr.MockResp{
					Data: []byte(`,,,`),
					Code: 200,
				},
			},
			true,
			"invalid character ',' looking for beginning of value",
		},
		{
			"badtoken",
			mr.MockRespList{
				mr.MockResp{
					Data: []byte(`{
						"description": "No authorization token provided.",
						"code": 401
					}`),
					Code: 401,
				},
			},
			true,
			"invalid token but no credentials provided",
		},
		{
			"clienterror",
			mr.MockRespList{
				mr.MockResp{
					Data: []byte{},
					Err:  errors.New("ka-boom"),
				},
			},
			true,
			"ka-boom",
		},
	}
	for _, tt := range tests {
		mresp.SetData(tt.responses)
		var err error
		t.Run(tt.name, func(t *testing.T) {
			if err = c.versionCheck(ctx); (err != nil) != tt.wantErr {
				t.Errorf("Client.versionCheck() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
		if !mresp.Empty() {
			t.Error("not all data in mock client consumed")
		}
		if tt.wantErr {
			assert.EqualError(t, err, tt.errstr)
		}
	}
}
