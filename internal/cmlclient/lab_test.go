package cmlclient

import (
	"io/ioutil"
	"testing"

	mr "github.com/rschmied/terraform-provider-cml2/m/v2/internal/mockresponder"
	"github.com/stretchr/testify/assert"
)

var (
	demoLab = []byte(`{
		"state": "STOPPED",
		"created": "2022-05-11T20:36:15+00:00",
		"modified": "2022-05-11T21:23:28+00:00",
		"lab_title": "vlandrop",
		"lab_description": "",
		"lab_notes": "",
		"owner": "00000000-0000-4000-a000-000000000000",
		"owner_username": "admin",
		"node_count": 2,
		"link_count": 1,
		"id": "labuuid",
		"groups": []
	}`)
	ownerUser = []byte(`{
		"id": "00000000-0000-4000-a000-000000000000",
		"created": "2022-04-29T13:44:46+00:00",
		"modified": "2022-05-20T10:57:42+00:00",
		"username": "admin",
		"fullname": "",
		"email": "",
		"description": "",
		"admin": true,
		"directory_dn": "",
		"groups": [],
		"labs": ["lab1"]
	}`)
	links = []byte(`["link1"]`)
	nodes = []byte(`["node1","node2"]`)
	node1 = []byte(`{
		"id": "node1",
		"lab_id": "lab1",
		"label": "alpine-0",
		"node_definition": "alpine",
		"state": "STOPPED"
	}`)
	node2 = []byte(`{
		"id": "node2",
		"lab_id": "lab1",
		"label": "alpine-1",
		"node_definition": "alpine",
		"state": "STOPPED"
	}`)
	ifacesn1  = []byte(`["n1i1"]`)
	ifacesn2  = []byte(`["n2i1"]`)
	ifacen1i1 = []byte(`{
		"id": "n1i1",
		"lab_id": "lab1",
		"node": "node1",
		"label": "eth0",
		"slot": 0,
		"type": "physical",
		"mac_address": "52:54:00:0c:e0:69",
		"is_connected": true,
		"state": "STOPPED"
	}`)
	ifacen2i1 = []byte(`{
		"id": "n2i1",
		"lab_id": "lab1",
		"node": "node2",
		"label": "eth0",
		"slot": 0,
		"type": "physical",
		"mac_address": "52:54:00:0c:e0:70",
		"is_connected": true,
		"state": "STOPPED"
	}`)
	linkn1n2 = []byte(`{
		"id": "link1",
		"interface_a": "n1i1",
		"interface_b": "n2i1",
		"lab_id": "lab1",
		"label": "alpine-0-eth0<->alpine-1-eth0",
		"link_capture_key": "",
		"node_a": "node1",
		"node_b": "node2",
		"state": "DEFINED_ON_CORE"
	}`)
)

func TestClient_GetLab(t *testing.T) {
	c := NewClient("https://bla.bla", true)
	mclient, ctx := mr.NewMockResponder()
	c.httpClient = mclient
	c.authChecked = true

	tests := []struct {
		name      string
		responses mr.MockRespList
		wantErr   bool
	}{
		{
			"lab1",
			mr.MockRespList{
				mr.MockResp{Data: []byte(`{"version": "2.4.1","ready": true}`)},
				mr.MockResp{Data: demoLab},
				mr.MockResp{Data: links, URL: `/links$`},
				mr.MockResp{Data: []byte(`{}`), URL: `/layer3_addresses$`},
				mr.MockResp{Data: ownerUser, URL: `/users/.+$`},
				mr.MockResp{Data: nodes, URL: `/nodes$`},
				mr.MockResp{Data: node1, URL: `/nodes/node1$`},
				mr.MockResp{Data: node2, URL: `/nodes/node2$`},
				mr.MockResp{Data: ifacesn1, URL: `/node1/interfaces$`},
				mr.MockResp{Data: ifacesn2, URL: `/node2/interfaces$`},
				mr.MockResp{Data: ifacen1i1, URL: `/interfaces/n1i1$`},
				mr.MockResp{Data: ifacen2i1, URL: `/interfaces/n2i1$`},
				mr.MockResp{Data: linkn1n2, URL: `/links/link1$`},
			},
			false,
		},
		{
			"incompatible controller",
			mr.MockRespList{
				mr.MockResp{
					Data: []byte(`{"version": "2.5.1","ready": true}`),
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
			lab, err := c.GetLab(ctx, "qweaa", false)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("Client.GetLab() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			assert.NotNil(t, lab)
			assert.Len(t, lab.Links, 1)
			assert.Len(t, lab.Nodes, 2)
		})
		if !mclient.Empty() {
			t.Error("not all data in mock client consumed")
		}
	}
}
func TestClient_GetLab_shallow(t *testing.T) {
	c := NewClient("https://bla.bla", true)
	mclient, ctx := mr.NewMockResponder()
	c.httpClient = mclient
	c.authChecked = true

	tests := []struct {
		name      string
		responses mr.MockRespList
		wantErr   bool
	}{
		{
			"good",
			mr.MockRespList{
				mr.MockResp{Data: []byte(`{"version": "2.4.1","ready": true}`)},
				mr.MockResp{Data: demoLab},
			},
			false,
		},
	}
	for _, tt := range tests {
		// enforce version check
		c.versionChecked = false
		mclient.SetData(tt.responses)
		t.Run(tt.name, func(t *testing.T) {
			lab, err := c.GetLab(ctx, "qweaa", true)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("Client.GetLab() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			assert.NotNil(t, lab)
		})
		if !mclient.Empty() {
			t.Error("not all data in mock client consumed")
		}
	}
}

func TestClient_ImportLab(t *testing.T) {
	c := NewClient("https://bla.bla", true)
	mclient, ctx := mr.NewMockResponder()
	c.httpClient = mclient
	c.authChecked = true
	c.versionChecked = true

	testfile := "testdata/labimport/twonodes.yaml"
	labyaml, err := ioutil.ReadFile(testfile)
	if err != nil {
		t.Errorf("Client.ImportLab() can't read testfile %s", testfile)
	}

	tests := []struct {
		name      string
		labyaml   string
		responses mr.MockRespList
		wantErr   bool
	}{
		{
			"good import",
			string(labyaml),
			// the import will also fetch the entire lab (not shallow!)
			mr.MockRespList{
				mr.MockResp{Data: []byte(`{"id": "lab-id-uuid", "warnings": [] }`)},
				mr.MockResp{Data: demoLab},
				// these responses are needed for not shallow...
				mr.MockResp{Data: links, URL: `/links$`},
				mr.MockResp{Data: []byte(`{}`), URL: `/layer3_addresses$`},
				mr.MockResp{Data: ownerUser, URL: `/users/.+$`},
				mr.MockResp{Data: nodes, URL: `/nodes$`},
				mr.MockResp{Data: node1, URL: `/nodes/node1$`},
				mr.MockResp{Data: node2, URL: `/nodes/node2$`},
				mr.MockResp{Data: ifacesn1, URL: `/node1/interfaces$`},
				mr.MockResp{Data: ifacesn2, URL: `/node2/interfaces$`},
				mr.MockResp{Data: ifacen1i1, URL: `/interfaces/n1i1$`},
				mr.MockResp{Data: ifacen2i1, URL: `/interfaces/n2i1$`},
				mr.MockResp{Data: linkn1n2, URL: `/links/link1$`},
			},
			false,
		},
		{
			"bad import",
			",,,", // invalid YAML
			mr.MockRespList{
				mr.MockResp{
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
	mclient, ctx := mr.NewMockResponder()
	c.httpClient = mclient
	c.apiToken = "expiredbadtoken"
	c.userpass = userPass{} // no password provided

	data := mr.MockRespList{
		mr.MockResp{
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
