package cmlclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sort"
)

// {
//   "id": "00000000-0000-4000-a000-000000000000",
//   "created": "2023-02-06T09:57:56+00:00",
//   "modified": "2023-03-26T11:18:59+00:00",
//   "username": "admin",
//   "fullname": "",
//   "email": "",
//   "description": "",
//   "admin": true,
//   "directory_dn": "",
//   "groups": [],
//   "labs": [
//     "fa3269a7-1357-472a-96ed-1c000883530d",
//     "3d5251b0-0455-408b-9c41-99fb29cc3bf1",
//   ],
//   "opt_in": true,
//   "resource_pool": null
// }

type User struct {
	ID           string   `json:"id,omitempty"`
	Created      string   `json:"created,omitempty"`
	Modified     string   `json:"modified,omitempty"`
	Username     string   `json:"username"`
	Password     string   `json:"password"`
	Fullname     string   `json:"fullname"`
	Email        string   `json:"email"`
	Description  string   `json:"description"`
	IsAdmin      bool     `json:"admin"`
	DirectoryDN  string   `json:"directory_dn,omitempty"`
	Groups       []string `json:"groups,omitempty"`
	Labs         []string `json:"labs,omitempty"`
	OptIn        bool     `json:"opt_in"`                  // with 2.5.0
	ResourcePool *string  `json:"resource_pool,omitempty"` // with 2.5.0
}

type UserList []*User

type userPatchPostAlias struct {
	Username     string   `json:"username"`
	Password     string   `json:"password,omitempty"`
	Fullname     string   `json:"fullname"`
	Email        string   `json:"email"`
	Description  string   `json:"description"`
	IsAdmin      bool     `json:"admin"`
	Groups       []string `json:"groups"`
	Labs         []string `json:"labs,omitempty"`          // can't be set
	OptIn        bool     `json:"opt_in"`                  // with 2.5.0
	ResourcePool *string  `json:"resource_pool,omitempty"` // with 2.5.0
}

func newUserAlias(user *User) userPatchPostAlias {
	upp := userPatchPostAlias{}

	upp.Username = user.Username
	upp.Password = user.Password
	upp.Fullname = user.Fullname
	upp.Email = user.Email
	upp.Description = user.Description
	upp.IsAdmin = user.IsAdmin
	upp.OptIn = user.OptIn
	upp.Groups = user.Groups
	upp.Labs = user.Labs
	upp.ResourcePool = user.ResourcePool

	return upp
}

// UserGet returns the user with the given `id`.
func (c *Client) UserGet(ctx context.Context, id string) (*User, error) {
	api := fmt.Sprintf("users/%s", id)
	user := &User{}
	err := c.jsonGet(ctx, api, user, 0)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// UserByName returns the user with the given username `name`.
func (c *Client) UserByName(ctx context.Context, name string) (*User, error) {
	api := fmt.Sprintf("users/%s/id", name)
	var userID string
	err := c.jsonGet(ctx, api, &userID, 0)
	if err != nil {
		return nil, err
	}
	return c.UserGet(ctx, userID)
}

// Users retrieves the list of all users which exist on the controller.
func (c *Client) Users(ctx context.Context) (UserList, error) {
	users := UserList{}
	err := c.jsonGet(ctx, "users", &users, 0)
	if err != nil {
		return nil, err
	}
	// sort the user list by their ID
	sort.Slice(users, func(i, j int) bool {
		return users[i].ID > users[j].ID
	})
	return users, nil
}

// UserDestroy deletes the user identified by the `id` (a UUIDv4).
func (c *Client) UserDestroy(ctx context.Context, id string) error {
	return c.jsonDelete(ctx, fmt.Sprintf("users/%s", id), 0)
}

// UserCreate creates a new user on the controller based on the data provided
// in the passed user parameter.
func (c *Client) UserCreate(ctx context.Context, user *User) (*User, error) {
	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(user)
	if err != nil {
		return nil, err
	}
	result := User{}
	err = c.jsonPost(ctx, "users", buf, &result, 0)
	if err != nil {
		return nil, err
	}
	return c.UserGet(ctx, result.ID)
}

// UserUpdate updates the given user which must exist.
func (c *Client) UserUpdate(ctx context.Context, user *User) (*User, error) {
	patchAlias := newUserAlias(user)
	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(patchAlias)
	if err != nil {
		return nil, err
	}
	result := User{}
	err = c.jsonPatch(ctx, fmt.Sprintf("users/%s", user.ID), buf, &result, 0)
	if err != nil {
		return nil, err
	}
	return c.UserGet(ctx, result.ID)
}

// UserGroups retrieves the list of all groups the user belongs to.
func (c *Client) UserGroups(ctx context.Context, id string) (GroupList, error) {
	api := fmt.Sprintf("users/%s/groups", id)
	idList := []string{}
	err := c.jsonGet(ctx, api, &idList, 0)
	if err != nil {
		return nil, err
	}

	groups := GroupList{}
	for _, id := range idList {
		group, err := c.GroupGet(ctx, id)
		if err != nil {
			return nil, err
		}
		groups = append(groups, group)
	}

	// sort the user list by their ID
	// (groups are a set so sorting is only done for test stability)
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].ID > groups[j].ID
	})
	return groups, nil
}
