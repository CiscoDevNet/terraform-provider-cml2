package cmlclient

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// {
// 	"username": "admin",
// 	"id": "90f84e38-a71c-4d57-8d90-00fa8a197385",
// 	"token": "123.123.123jwtdata",
// 	"admin": false
// }

type Auth struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Token    string `json:"token"`
	Admin    bool   `json:"admin"`
}

type userPass struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (up userPass) valid() bool {
	return len(up.Username) > 0 && len(up.Password) > 0
}

// technically, authokAPI requires auth, but it's used specifically
// to test whether auth is OK, so it will take a different path
func (c *Client) authRequired(api *url.URL) bool {
	url := api.String()
	return !(strings.HasSuffix(url, authAPI) ||
		strings.HasSuffix(url, authokAPI) ||
		strings.HasSuffix(url, systeminfoAPI))
}

func (c *Client) authenticate(ctx context.Context, userpass userPass, depth int32) error {
	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(userpass)
	if err != nil {
		return err
	}
	auth := &Auth{}
	err = c.jsonPost(ctx, authAPI, buf, auth, depth)
	if err != nil {
		return err
	}
	log.Printf("user id %s, is admin: %s", auth.ID, strconv.FormatBool(auth.Admin))
	c.apiToken = auth.Token
	return nil
}

// SetToken sets a specific API token to be used. A token takes precedence over
// a username/password. However, if the token expires, the username/password are
// used to authorize the client again. An error is raised if no token and no
// username/password are provided or if the token expires when no username/password
// are set.
func (c *Client) SetToken(token string) {
	c.apiToken = token
}

// SetUsernamePassword sets the username and the password to be used with the
// client for all authentications.
func (c *Client) SetUsernamePassword(username, password string) {
	c.userpass = userPass{
		username, password,
	}
}

// SetCACert sets a specific X.509 CA certificate to use with the client.
// If no cert is set, the system trust anchors are used for cert verification.
func (c *Client) SetCACert(cert []byte) error {
	caCertPool := x509.NewCertPool()
	ok := caCertPool.AppendCertsFromPEM(cert)
	if !ok {
		return errors.New("failed to parse root certificate")
	}
	httpClient, ok := c.httpClient.(*http.Client)
	if !ok {
		return errors.New("can't set certs on mocked client")
	}
	tr := httpClient.Transport.(*http.Transport)
	tr.TLSClientConfig.RootCAs = caCertPool
	return nil
}
