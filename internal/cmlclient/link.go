package cmlclient

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
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

type linkAlias struct {
	ID      string `json:"id"`
	State   string `json:"state"`
	SrcID   string `json:"interface_a"`
	DstID   string `json:"interface_b"`
	SrcNode string `json:"node_a"`
	DstNode string `json:"node_b"`
	Label   string `json:"label"`
	PCAPkey string `json:"link_capture_key"`
}

type Link struct {
	ID      string `json:"id"`
	State   string `json:"state"`
	Label   string `json:"label"`
	PCAPkey string `json:"link_capture_key"`

	ifaceA *Interface
	ifaceB *Interface
}

func (l Link) MarshalJSON() ([]byte, error) {
	link := linkAlias{
		ID:      l.ID,
		State:   l.State,
		SrcID:   l.ifaceA.ID,
		DstID:   l.ifaceB.ID,
		SrcNode: l.ifaceA.node.ID,
		DstNode: l.ifaceB.node.ID,
		Label:   l.Label,
		PCAPkey: l.PCAPkey,
	}
	return json.Marshal(link)
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
	err := c.jsonGet(ctx, api, linkIDlist)
	if err != nil {
		return nil, err
	}
	return *linkIDlist, nil
}

func (c *Client) getLinksForLab(ctx context.Context, lab *Lab, linkIDlist IDlist) error {
	// api := fmt.Sprintf("labs/%s/links", lab.ID)
	// linkIDlist := &IDlist{}
	// err := c.jsonGet(ctx, api, linkIDlist)
	// if err != nil {
	// 	return err
	// }

	linkList := linkList{}
	for _, linkID := range linkIDlist {
		api := fmt.Sprintf("labs/%s/links/%s", lab.ID, linkID)
		link := &linkAlias{}
		err := c.jsonGet(ctx, api, link)
		if err != nil {
			return err
		}
		realLink := &Link{
			ID:      link.ID,
			State:   link.State,
			Label:   link.Label,
			PCAPkey: link.PCAPkey,
			ifaceA:  c.findInterface(lab.Nodes, link.SrcID),
			ifaceB:  c.findInterface(lab.Nodes, link.DstID),
		}
		linkList = append(linkList, realLink)
	}
	lab.Links = linkList
	return nil
}
