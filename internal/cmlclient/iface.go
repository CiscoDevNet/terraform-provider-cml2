package cmlclient

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
)

// {
// 	"id": "20681832-36e8-4ba9-9d8d-0588e0f7b517",
// 	"lab_id": "52d5c824-e10c-450a-b9c5-b700bd3bc17a",
// 	"node": "9efb1503-7e2a-4d2a-959e-865209f1acc0",
// 	"label": "port",
// 	"slot": 0,
// 	"type": "physical",
// 	"device_name": "",
// 	"dst_udp_port": null,
// 	"src_udp_port": null,
// 	"mac_address": null,
// 	"is_connected": true,
// 	"state": "STARTED"
// }

type Interface struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Type        string `json:"type"`
	Slot        int    `json:"slot"`
	State       string `json:"state"`
	MACaddress  string `json:"mac_address"`
	IsConnected bool   `json:"is_connected"`
	DeviceName  string `json:"device_name"`

	// extra
	IP4  []string `json:"ip4"`
	IP6  []string `json:"ip6"`
	node *Node
}

func (imap interfaceMap) MarshalJSON() ([]byte, error) {
	ilist := []*Interface{}
	for _, iface := range imap {
		ilist = append(ilist, iface)
	}
	// we want this as a stable sort by interface UUID
	sort.Slice(ilist, func(i, j int) bool {
		return ilist[i].ID < ilist[j].ID
	})
	return json.Marshal(ilist)
}

func (c *Client) getInterfacesForNode(ctx context.Context, id string, node *Node) error {
	api := fmt.Sprintf("labs/%s/nodes/%s/interfaces", id, node.ID)
	interfaceIDlist := &IDlist{}
	err := c.jsonGet(ctx, api, interfaceIDlist)
	if err != nil {
		return err
	}

	interfaceMap := make(interfaceMap)
	for _, ifaceID := range *interfaceIDlist {
		api = fmt.Sprintf("labs/%s/interfaces/%s", id, ifaceID)
		iface := &Interface{node: node}
		err := c.jsonGet(ctx, api, iface)
		if err != nil {
			return err
		}
		interfaceMap[ifaceID] = iface
	}
	node.Interfaces = interfaceMap
	return nil
}

func (c *Client) findInterface(nodes nodeMap, id string) *Interface {
	for _, node := range nodes {
		if iface, found := node.Interfaces[id]; found {
			return iface
		}
	}
	return nil
}

// {
// 	"00da52b6-2683-49c0-ba3a-ace877dea4ca": {
// 	  "name": "alpine-0",
// 	  "interfaces": {
// 		"52:54:00:00:00:09": {
// 		  "id": "3b45184f-7041-4300-aef2-2b97d8e763a8",
// 		  "label": "eth0",
// 		  "ip4": [
// 			"192.168.122.35"
// 		  ],
// 		  "ip6": [
// 			"fe80::5054:ff:fe00:9"
// 		  ]
// 		}
// 	  }
// 	},
// 	"0df7a717-9826-4729-9fe1-bc4932498c83": {
// 	  "name": "alpine-1",
// 	  "interfaces": {
// 		"52:54:00:00:00:08": {
// 		  "id": "6bec8956-f812-4fb3-9551-aef4410807ec",
// 		  "label": "eth0",
// 		  "ip4": [
// 			"192.168.122.34"
// 		  ],
// 		  "ip6": [
// 			"fe80::5054:ff:fe00:8"
// 		  ]
// 		}
// 	  }
// 	}
// }

type l3nodes map[string]*l3node

type l3node struct {
	Name       string                 `json:"name"`
	Interfaces map[string]l3interface `json:"interfaces"`
}

type l3interface struct {
	ID    string   `json:"id"`
	Label string   `json:"label"`
	IP4   []string `json:"ip4"`
	IP6   []string `json:"ip6"`
}

func (c *Client) getL3Info(ctx context.Context, id string) (*l3nodes, error) {
	api := fmt.Sprintf("labs/%s/layer3_addresses", id)
	l3n := &l3nodes{}
	err := c.jsonGet(ctx, api, l3n)
	if err != nil {
		return nil, err
	}
	return l3n, nil
}
