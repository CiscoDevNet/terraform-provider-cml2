package cmlclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sort"
)

const (
	LinkStateDefined = "DEFINED_ON_CORE"
	LinkStateStopped = "STOPPED"
	LinkStateStarted = "STARTED"
)

// {
// 	"id": "4d76f475-2915-444e-bfd1-425a517120bc",
// 	"interface_a": "20681832-36e8-4ba9-9d8d-0588e0f7b517",
// 	"interface_b": "1959cc9f-361c-410e-a960-9d9a896482a0",
// 	"lab_id": "52d5c824-e10c-450a-b9c5-b700bd3bc17a",
// 	"label": "ext-conn-0-port<->unmanaged-switch-0-port0",
// 	"link_capture_key": "d827ce92-db2e-4933-bc0d-7a2c38e39ad5",
// 	"node_a": "9efb1503-7e2a-4d2a-959e-865209f1acc0",
// 	"node_b": "1cc0cbcd-6b4f-4bbe-9f69-2c3da5e3495a",
// 	"state": "STARTED"
// }

// Link defines the data structure for a CML link between nodes.
type Link struct {
	ID      string `json:"id"`
	LabID   string `json:"lab_id"`
	State   string `json:"state"`
	Label   string `json:"label"`
	PCAPkey string `json:"link_capture_key"`
	SrcID   string `json:"interface_a"`
	DstID   string `json:"interface_b"`
	SrcNode string `json:"node_a"`
	DstNode string `json:"node_b"`
	SrcSlot int    `json:"slot_a"`
	DstSlot int    `json:"slot_b"`

	// not exported, needed for internal linking
	ifaceA *Interface
	ifaceB *Interface
}

func (llist linkList) MarshalJSON() ([]byte, error) {
	type slist linkList
	newlist := slist(llist)
	// we want this as a stable sort by link UUID
	sort.Slice(newlist, func(i, j int) bool {
		return newlist[i].ID < newlist[j].ID
	})
	return json.Marshal(newlist)
}

func (c *Client) getLinkIDsForLab(ctx context.Context, lab *Lab) (IDlist, error) {
	api := fmt.Sprintf("labs/%s/links", lab.ID)
	linkIDlist := &IDlist{}
	err := c.jsonGet(ctx, api, linkIDlist, 0)
	if err != nil {
		return nil, err
	}
	return *linkIDlist, nil
}

func (c *Client) getLinksForLab(ctx context.Context, lab *Lab, linkIDlist IDlist) error {
	linkList := linkList{}
	for _, linkID := range linkIDlist {
		api := fmt.Sprintf("labs/%s/links/%s", lab.ID, linkID)
		link := &Link{}
		err := c.jsonGet(ctx, api, link, 0)
		if err != nil {
			return err
		}
		link.LabID = lab.ID
		linkList = append(linkList, link)
	}
	lab.Links = linkList
	return nil
}

// LinkGet returns the link data for the given `labID` and `linkID`. If `deep` is
// set to `true` then bot interface and node data for the given link are also
// fetched from the controller.
func (c *Client) LinkGet(ctx context.Context, labID, linkID string, deep bool) (*Link, error) {
	api := fmt.Sprintf("labs/%s/links/%s", labID, linkID)
	link := &Link{}
	err := c.jsonGet(ctx, api, link, 0)
	if err != nil {
		return nil, err
	}

	link.LabID = labID

	if deep {
		var (
			err error
			// ifaceA, ifaceB *Interface
		)

		ifaceA := &Interface{
			ID:    link.SrcID,
			LabID: labID,
			Node:  link.SrcNode,
			node:  &Node{ID: link.SrcNode, LabID: labID},
		}
		ifaceA, err = c.InterfaceGet(ctx, ifaceA)
		if err != nil {
			return nil, err
		}
		ifaceA.node, err = c.NodeGet(ctx, ifaceA.node, false)
		if err != nil {
			return nil, err
		}

		ifaceB := &Interface{
			ID:    link.DstID,
			LabID: labID,
			Node:  link.DstNode,
			node:  &Node{ID: link.DstNode, LabID: labID},
		}
		ifaceB, err = c.InterfaceGet(ctx, ifaceB)
		if err != nil {
			return nil, err
		}
		ifaceB.node, err = c.NodeGet(ctx, ifaceB.node, false)
		if err != nil {
			return nil, err
		}

		link.ifaceA = ifaceA
		link.ifaceB = ifaceB
		link.SrcSlot = ifaceA.Slot
		link.DstSlot = ifaceB.Slot
	}
	return link, err
}

// LinkCreate creates a link based on the the data passed in `link`. Required fields
// are the `LabID` and either a pair of interfaces `SrcID` / `DstID` or a pair of
// nodes `SrcNode` / `DstNode`. With nodes it's also possible to provide specific
// slots in `SrcSlot` / `DstSlot` where the link should be created.
// If one or both of the provided slots aren't available, then new interfaces will
// be craeted. If interface creation fails or the provided Interface IDs can't be
// found, the API returns an error, otherwise the returned Link variable has the
// updated link data.
// Node: -1 for a slot means: use next free slot. Specific slots run from 0 to the
// maximum slot number -1 per the node definition of the node type.
func (c *Client) LinkCreate(ctx context.Context, link *Link) (*Link, error) {
	api := fmt.Sprintf("labs/%s/links", link.LabID)

	var (
		err          error
		nodeA, nodeB *Node
	)

	if len(link.SrcNode) > 0 && len(link.DstNode) > 0 {

		nodeA = &Node{LabID: link.LabID, ID: link.SrcNode}
		// if c.useCache {
		// 	nodeA, err = c.NodeGet(ctx, nodeA, false)
		// 	if err != nil {
		// 		return nil, err
		// 	}
		// }
		err = c.getInterfacesForNode(ctx, nodeA)
		if err != nil {
			return nil, err
		}

		nodeB = &Node{LabID: link.LabID, ID: link.DstNode}
		// if c.useCache {
		// 	nodeB, err = c.NodeGet(ctx, nodeB, false)
		// 	if err != nil {
		// 		return nil, err
		// 	}
		// }
		err = c.getInterfacesForNode(ctx, nodeB)
		if err != nil {
			return nil, err
		}

		matches := func(slot int, iface *Interface) bool {
			if !iface.IsPhysical() {
				return false
			}
			if slot >= 0 {
				if iface.Slot == slot && !iface.IsConnected {
					return true
				}
			} else {
				if !iface.IsConnected {
					return true
				}
			}
			return false
		}

		for _, iface := range nodeA.Interfaces {
			if matches(link.SrcSlot, iface) {
				iface.IsConnected = true
				link.ifaceA = iface
				link.SrcID = iface.ID
				break
			}
		}

		for _, iface := range nodeB.Interfaces {
			if matches(link.DstSlot, iface) {
				iface.IsConnected = true
				link.ifaceA = iface
				link.DstID = iface.ID
				break
			}
		}

		if len(link.SrcID) == 0 {
			iface, err := c.InterfaceCreate(ctx, link.LabID, link.SrcNode, link.SrcSlot)
			if err != nil {
				return nil, err
			}
			iface.IsConnected = true
			link.SrcID = iface.ID
			link.ifaceA = iface
		}
		if len(link.DstID) == 0 {
			iface, err := c.InterfaceCreate(ctx, link.LabID, link.DstNode, link.DstSlot)
			if err != nil {
				return nil, err
			}
			iface.IsConnected = true
			link.DstID = iface.ID
			link.ifaceB = iface
		}
	}

	newLink := struct {
		SrcInt string `json:"src_int"`
		DstInt string `json:"dst_int"`
	}{
		SrcInt: link.SrcID,
		DstInt: link.DstID,
	}

	buf := &bytes.Buffer{}
	err = json.NewEncoder(buf).Encode(newLink)
	if err != nil {
		return nil, err
	}

	newLinkResult := struct {
		ID string `json:"id"`
	}{}
	err = c.jsonPost(ctx, api, buf, &newLinkResult, 0)
	if err != nil {
		return nil, err
	}

	return c.LinkGet(ctx, link.LabID, newLinkResult.ID, true)
}
