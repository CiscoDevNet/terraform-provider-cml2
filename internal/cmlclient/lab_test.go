package cmlclient

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"reflect"
	"testing"

	mc "github.com/rschmied/terraform-provider-cml2/m/v2/internal/mockclient"
	"github.com/stretchr/testify/assert"
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
	"node_count": 2,
	"link_count": 1,
	"id": "52d5c824-e10c-450a-b9c5-b700bd3bc17a",
	"groups": []
  }`)

func TestClient_GetLab(t *testing.T) {
	c := NewClient("https://bla.bla", true)
	mclient, ctx := mc.NewMockClient()
	c.httpClient = mclient
	c.authChecked = true

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
					t.Errorf("Client.GetLab() error = %v, wantErr %v", err, tt.wantErr)
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

func TestClient_ImportLab(t *testing.T) {
	c := NewClient("https://bla.bla", true)
	mclient, ctx := mc.NewMockClient()
	c.httpClient = mclient
	c.authChecked = true
	c.versionChecked = true

	// additional data read for the lab import
	labdetails, err := ioutil.ReadFile("testdata/labimport/labs-lab-id-uuid.json")
	if err != nil {
		t.Errorf("Client.ImportLab() can't read testfile %s", labdetails)
	}

	testfile := "testdata/labimport/twonodes.yaml"
	labyaml, err := ioutil.ReadFile(testfile)
	if err != nil {
		t.Errorf("Client.ImportLab() can't read testfile %s", testfile)
	}

	tests := []struct {
		name      string
		labyaml   string
		responses mc.MockRespList
		wantErr   bool
	}{
		{
			"good import",
			string(labyaml),
			mc.MockRespList{
				mc.MockResp{
					Data: []byte(`{"id": "lab-id-uuid", "warnings": [] }`),
					Code: 200,
				},
				mc.MockResp{
					Data: labdetails,
					Code: 200,
				},
			},
			false,
		},
		{
			"bad import",
			",,,", // invalid YAML
			mc.MockRespList{
				mc.MockResp{
					Data: []byte(`{
					"description": "Bad request: while parsing a block node\nexpected the node content, but found ','\n  in \"<unicode string>\", line 1, column 1:\n    ,,,\n    ^.",
					"code": 400}
					`),
					Code: 400,
				},
			},
			true,
		},
	}

	for _, tt := range tests {
		mclient.SetData(tt.responses)
		t.Run(tt.name, func(t *testing.T) {
			lab, err := c.ImportLab(ctx, tt.labyaml)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("Client.GetLab() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			assert.NotNil(t, lab)
			// TODO: when adding more tests, the node/link count needs to be
			// parametrized!!
			assert.Equal(t, lab.NodeCount, 2)
			assert.Equal(t, lab.LinkCount, 1)
		})
		if !mclient.Empty() {
			t.Error("not all data in mock client consumed")
		}
	}
}

func TestClient_ImportLabBadAuth(t *testing.T) {
	c := NewClient("https://bla.bla", true)
	mclient, ctx := mc.NewMockClient()
	c.httpClient = mclient
	c.apiToken = "expiredbadtoken"
	c.userpass = userPass{} // no password provided

	data := mc.MockRespList{
		mc.MockResp{
			Data: []byte(`{
				"description": "description": "401: Unauthorized",
				"code":        401
			}`),
			Code: 401,
		},
	}
	mclient.SetData(data)
	lab, err := c.ImportLab(ctx, `{}`)

	if !mclient.Empty() {
		t.Error("not all data in mock client consumed")
	}

	assert.NotNil(t, err)
	assert.Nil(t, lab)
}
