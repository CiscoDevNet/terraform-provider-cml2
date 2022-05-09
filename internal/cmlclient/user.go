package cmlclient

import (
	"context"
	"fmt"
)

// {
// 	"id": "00000000-0000-4000-a000-000000000000",
// 	"created": "2022-04-29T13:44:46+00:00",
// 	"modified": "2022-05-05T16:16:40+00:00",
// 	"username": "admin",
// 	"fullname": "",
// 	"email": "",
// 	"description": "",
// 	"admin": true,
// 	"directory_dn": "",
// 	"groups": [],
// 	"labs": [
// 	  "52d5c824-e10c-450a-b9c5-b700bd3bc17a",
// 	  "b96c589f-c449-4013-a7a8-4ee57cbef025",
// 	  "c48222c2-cb2d-4255-8cc4-891bcd810014"
// 	]
// }

type User struct {
	ID          string `json:"id"`
	Created     string `json:"created"`
	Modified    string `json:"modified"`
	Username    string `json:"username"`
	Fullname    string `json:"fullname"`
	Email       string `json:"email"`
	Description string `json:"lab_description"`
	IsAdmin     bool   `json:"admin"`
	DirectoryDN string `json:"directory_dn"`
}

func (c *Client) getUser(ctx context.Context, id string) (*User, error) {
	api := fmt.Sprintf("users/%s", id)
	user := &User{}
	err := c.jsonGet(ctx, api, user)
	if err != nil {
		return nil, err
	}
	return user, nil
}
