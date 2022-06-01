package cmlclient

import (
	"bytes"
	"encoding/json"
	"net/http"
	"reflect"
	"testing"

	mr "github.com/rschmied/terraform-provider-cml2/m/v2/internal/mockresponder"
)

func TestClient_GetImageDefs(t *testing.T) {

	c := NewClient("https://bla.bla", true)

	mresp, ctx := mr.NewMockResponder()
	c.httpClient = mresp
	c.authChecked = true
	c.versionChecked = true

	tests := []struct {
		name      string
		responses mr.MockRespList
		wantErr   bool
	}{
		{
			"good",
			mr.MockRespList{
				mr.MockResp{
					Data: []byte(
						`[{
							"id": "alpine-3-10-base",
							"node_definition_id": "alpine",
							"description": "Alpine Linux and network tools",
							"label": "Alpine 3.10",
							"disk_image": "alpine-3-10-base.qcow2",
							"read_only": true,
							"ram": null,
							"cpus": null,
							"cpu_limit": null,
							"data_volume": null,
							"boot_disk_size": null,
							"disk_subfolder": "alpine-3-10-base",
							"schema_version": "0.0.1"
						}]`),
				},
			},
			false,
		},
		{
			"bad",
			mr.MockRespList{
				mr.MockResp{
					Data: []byte(`"something failed!`),
					Code: http.StatusInternalServerError,
				},
			},
			true,
		},
	}

	for _, tt := range tests {
		mresp.SetData(tt.responses)
		t.Run(tt.name, func(t *testing.T) {
			got, err := c.GetImageDefs(ctx)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("Client.GetImageDefs() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			expected := []ImageDefinition{}
			b := bytes.NewReader(mresp.LastData())
			err = json.NewDecoder(b).Decode(&expected)
			if err != nil {
				t.Error("bad test data")
				return
			}
			if !reflect.DeepEqual(got, expected) {
				t.Errorf("Client.GetImageDefs() = %v, want %v", got, expected)
			}
		})
		if !mresp.Empty() {
			t.Error("not all data in mock client consumed")
		}
	}
}
