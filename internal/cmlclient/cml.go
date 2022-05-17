package cmlclient

import (
	"crypto/tls"
	"net/http"
	"time"
)

type Client struct {
	httpClient *http.Client
	apiToken   string
	host       string
	userpass   userPass
}

func NewClient(host string, insecure bool) *Client {
	tr := http.DefaultTransport.(*http.Transport)
	tr.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: insecure,
	}

	return &Client{
		host:     host,
		apiToken: "",
		userpass: userPass{},
		httpClient: &http.Client{
			Timeout:   15 * time.Second,
			Transport: tr,
		},
	}
}
