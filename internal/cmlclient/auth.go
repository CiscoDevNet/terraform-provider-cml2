package cmlclient

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
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

func (c *Client) authenticate(ctx context.Context, userpass userPass) error {
	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(userpass)
	if err != nil {
		return err
	}
	auth := &Auth{}
	err = c.jsonPost(ctx, authAPI, buf, auth)
	if err != nil {
		return err
	}
	log.Printf("user id %s, is admin: %s", auth.ID, strconv.FormatBool(auth.Admin))
	c.apiToken = auth.Token
	return nil
}

func (c *Client) SetToken(token string) {
	c.apiToken = token
}

func (c *Client) SetUsernamePassword(username, password string) {
	c.userpass = userPass{
		username, password,
	}
}

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
