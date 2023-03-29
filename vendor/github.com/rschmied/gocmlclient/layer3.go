package cmlclient

import (
	"context"
	"fmt"
)

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
	err := c.jsonGet(ctx, api, l3n, 0)
	if err != nil {
		return nil, err
	}
	return l3n, nil
}
