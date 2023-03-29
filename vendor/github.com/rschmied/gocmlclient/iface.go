package cmlclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sort"
)

/*
{
	"id": "e87c811d-5459-4390-8e92-317bb9dc23e8",
	"lab_id": "024fa9f4-5e5e-4e94-9f85-29f147e09689",
	"node": "f902d112-2a93-4c9f-98e6-adea6dc16fef",
	"label": "eth0",
	"slot": 0,
	"type": "physical",
	"device_name": null,
	"dst_udp_port": 21001,
	"src_udp_port": 21000,
	"mac_address": "52:54:00:1e:af:9b",
	"is_connected": true,
	"state": "STARTED"
}
*/

const (
	IfaceStateDefined = "DEFINED_ON_CORE"
	IfaceStateStopped = "STOPPED"
	IfaceStateStarted = "STARTED"

	IfaceTypePhysical = "physical"
	IfaceTypeLoopback = "loopback"
)

type Interface struct {
	ID          string `json:"id"`
	LabID       string `json:"lab_id"`
	Node        string `json:"node"`
	Label       string `json:"label"`
	Slot        int    `json:"slot"`
	Type        string `json:"type"`
	DeviceName  string `json:"device_name"`
	SrcUDPport  int    `json:"src_udp_port"`
	DstUDPport  int    `json:"dst_udp_port"`
	MACaddress  string `json:"mac_address"`
	IsConnected bool   `json:"is_connected"`
	State       string `json:"state"`

	// extra
	IP4 []string `json:"ip4"`
	IP6 []string `json:"ip6"`

	// needed for internal linking
	node *Node
}

func (iface Interface) Exists() bool {
	return iface.State != IfaceStateDefined
}

func (iface Interface) Runs() bool {
	return iface.State == IfaceStateStarted
}

func (iface Interface) IsPhysical() bool {
	return iface.Type == IfaceTypePhysical
}

func (c *Client) updateCachedIface(existingIface, iface *Interface) *Interface {
	// this is a no-op at this point, we don't allow updating interfaces
	return existingIface
}

func (c *Client) cacheIface(iface *Interface, err error) (*Interface, error) {
	if !c.useCache || err != nil {
		return iface, err
	}

	c.mu.RLock()
	lab, ok := c.labCache[iface.LabID]
	c.mu.RUnlock()
	if !ok {
		return iface, err
	}

	c.mu.RLock()
	node, ok := lab.Nodes[iface.Node]
	c.mu.RUnlock()
	if !ok {
		return iface, err
	}
	c.mu.RLock()
	interfaces := node.Interfaces
	c.mu.RUnlock()
	for _, nodeIface := range interfaces {
		if nodeIface.ID == iface.ID {
			return c.updateCachedIface(nodeIface, iface), nil
		}
	}

	iface.node = node // internal linking
	c.mu.Lock()
	node.Interfaces = append(node.Interfaces, iface)
	c.mu.Unlock()
	return iface, nil
}

func (c *Client) getCachedIface(iface *Interface) (*Interface, bool) {
	if !c.useCache {
		return nil, false
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	lab, ok := c.labCache[iface.LabID]
	if !ok {
		return nil, false
	}

	node, ok := lab.Nodes[iface.Node]
	if !ok {
		return iface, false
	}

	for _, nodeIface := range node.Interfaces {
		if nodeIface != nil && nodeIface.ID == iface.ID {
			if nodeIface.node == nil {
				nodeIface.node = node
			}
			return nodeIface, true
		}
	}

	return iface, false
}

func (c *Client) deleteCachedIface(iface *Interface, err error) error {
	if !c.useCache || err != nil {
		return err
	}

	c.mu.RLock()
	lab, ok := c.labCache[iface.LabID]
	c.mu.RUnlock()
	if !ok {
		return err
	}

	c.mu.RLock()
	node, ok := lab.Nodes[iface.Node]
	c.mu.RUnlock()
	if !ok {
		return err
	}

	c.mu.Lock()
	newList := InterfaceList{}
	for _, nodeIface := range node.Interfaces {
		if nodeIface.ID != iface.ID {
			newList = append(newList, nodeIface)
		}
	}
	node.Interfaces = newList
	c.mu.Unlock()
	return nil
}

func (c *Client) getInterfacesForNode(ctx context.Context, node *Node) error {
	// with the data=true option, we get not only the list of IDs but the
	// interfaces themselves as well!
	api := fmt.Sprintf("labs/%s/nodes/%s/interfaces?data=true", node.LabID, node.ID)
	interfaceList := InterfaceList{}
	err := c.jsonGet(ctx, api, &interfaceList, 0)
	if err != nil {
		return err
	}

	// sort the interface list by slot
	sort.Slice(interfaceList, func(i, j int) bool {
		return interfaceList[i].Slot < interfaceList[j].Slot
	})
	for _, iface := range interfaceList {
		c.cacheIface(iface, nil)
	}
	c.mu.Lock()
	node.Interfaces = interfaceList
	c.mu.Unlock()
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

// InterfaceGet returns the interface identified by its `ID` (iface.ID).
func (c *Client) InterfaceGet(ctx context.Context, iface *Interface) (*Interface, error) {

	if iface, ok := c.getCachedIface(iface); ok {
		return iface, nil
	}

	api := fmt.Sprintf("labs/%s/interfaces/%s", iface.LabID, iface.ID)
	err := c.jsonGet(ctx, api, iface, 0)
	return c.cacheIface(iface, err)
}

// InterfaceCreate creates an interface in the given lab and node.  If the slot
// is >= 0, the request creates all unallocated slots up to and including
// that slot. Conversely, if the slot is < 0 (e.g. -1), the next free slot is used.
func (c *Client) InterfaceCreate(ctx context.Context, labID, nodeID string, slot int) (*Interface, error) {

	var slotPtr *int

	if slot >= 0 {
		slotPtr = &slot
	}

	newIface := struct {
		Node string `json:"node"`
		Slot *int   `json:"slot,omitempty"`
	}{
		Node: nodeID,
		Slot: slotPtr,
	}

	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(newIface)
	if err != nil {
		return nil, err
	}

	// This is quite awkward, not even sure if it's a good REST design practice:
	// "Returns a JSON object that identifies the interface that was created. In
	// the case of bulk interface creation, returns a JSON array of such
	// objects." <-- from the API documentation
	// A list is returned when slot is defined, even if it's just creating
	// one interface

	api := fmt.Sprintf("labs/%s/interfaces", labID)
	if slotPtr == nil {
		result := Interface{}
		err = c.jsonPost(ctx, api, buf, &result, 0)
		if err != nil {
			return nil, err
		}
		return c.cacheIface(&result, err)
	}

	// this is when a slot has been provided; the API provides now a list of
	// interfaces
	result := []Interface{}
	err = c.jsonPost(ctx, api, buf, &result, 0)
	if err != nil {
		return nil, err
	}

	lastIface := &result[len(result)-1]
	for _, li := range result {
		c.cacheIface(&li, nil)
	}
	return lastIface, nil
}
