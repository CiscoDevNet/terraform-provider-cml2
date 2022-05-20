package cmlclient

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"

	mc "github.com/rschmied/terraform-provider-cml2/m/v2/internal/mockclient"
)

func TestClient_GetImageDefs(t *testing.T) {

	c := NewClient("https://bla.bla", true)

	mclient, ctx := mc.NewMockClient()
	c.httpClient = mclient
	c.authChecked = true
	c.versionChecked = true

	tests := []struct {
		name      string
		responses mc.MockRespList
		wantErr   bool
	}{
		{
			"good",
			mc.MockRespList{
				mc.MockResp{
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
					Code: 200,
				},
			},
			false,
		},
	}

	for _, tt := range tests {
		mclient.SetData(tt.responses)
		t.Run(tt.name, func(t *testing.T) {
			got, err := c.GetImageDefs(ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.GetImageDefs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			expected := []ImageDefinition{}
			b := bytes.NewReader(mclient.LastData())
			err = json.NewDecoder(b).Decode(&expected)
			if err != nil {
				t.Error("bad test data")
				return
			}
			if !reflect.DeepEqual(got, expected) {
				t.Errorf("Client.GetImageDefs() = %v, want %v", got, expected)
			}
		})
		if !mclient.Empty() {
			t.Error("not all data in mock client consumed")
		}
	}
}
