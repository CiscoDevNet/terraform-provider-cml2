package cmlclient

import (
	"bytes"
	"context"
	"encoding/json"
	"reflect"
	"testing"

	mc "github.com/rschmied/terraform-provider-cml2/m/v2/internal/mockclient"
)

var shallowLab = []byte(`{
	"state": "DEFINED_ON_CORE",
	"created": "2022-04-29T14:17:24+00:00",
	"modified": "2022-05-04T16:43:48+00:00",
	"lab_title": "demobla",
	"lab_description": "",
	"lab_notes": "",
	"owner": "00000000-0000-4000-a000-000000000000",
	"owner_username": "admin",
	"node_count": 4,
	"link_count": 3,
	"id": "52d5c824-e10c-450a-b9c5-b700bd3bc17a",
	"groups": []
  }`)

func TestClient_GetLab(t *testing.T) {

	ctx := context.TODO()

	c := NewClient("https://bla.bla", true)
	mclient, ctx := mc.NewMockClient()
	c.httpClient = mclient
	c.SetToken("blabla")

	tests := []struct {
		name      string
		responses mc.MockRespList
		wantErr   bool
	}{
		{
			"lab1",
			mc.MockRespList{
				mc.MockResp{
					Data: []byte(`{"version": "2.4.1","ready": true}`),
					Code: 200,
				},
				mc.MockResp{Data: shallowLab, Code: 200},
			},
			false,
		},
		{
			"incompatible controller",
			mc.MockRespList{
				mc.MockResp{
					Data: []byte(`{"version": "2.5.1","ready": true}`),
					Code: 200,
				},
			},
			true,
		},
	}
	for _, tt := range tests {
		// enforce version check
		c.versionChecked = false
		mclient.SetData(tt.responses)
		t.Run(tt.name, func(t *testing.T) {
			lab, err := c.GetLab(ctx, "qwe", true)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("Client.VersionCheck() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			expected := &labAlias{}
			b := bytes.NewReader([]byte(shallowLab))
			err = json.NewDecoder(b).Decode(expected)
			if err != nil {
				t.Errorf("bad test data %s", err)
				return
			}
			if !reflect.DeepEqual(lab, &(expected.Lab)) {
				t.Errorf("Client.GetLab() = %+v, want %+v", lab, expected.Lab)
			}

		})
		if !mclient.Empty() {
			t.Error("not all data in mock client consumed")
		}
	}
}
