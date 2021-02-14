package thingscloud

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

const (
	// APIEndpoint is the public culturedcode https endpoint
	APIEndpoint = "https://cloud.culturedcode.com/version/1"
)

var (
	// ErrUnauthorized is returned by the API when the credentials are wrong
	ErrUnauthorized = errors.New("unauthorized")
)

// Client is a culturedcode cloud client. It can be used to interact with the
// things cloud to manage your data.
type Client struct {
	Endpoint string
	EMail    string
	password string

	client *http.Client
	common service

	Accounts *AccountService
}

type service struct {
	client *Client
}

// New initializes a things client
func New(endpoint, email, password string) *Client {
	c := &Client{
		Endpoint: endpoint,
		EMail:    email,
		password: password,

		client: &http.Client{},
	}
	c.common.client = c
	c.Accounts = (*AccountService)(&c.common)
	return c
}

// ThingsUserAgent is the http user-agent header set by things for mac Version 3.13.8 (31308504)
const ThingsUserAgent = "ThingsMac/31308504mas"

func (c *Client) do(req *http.Request) (*http.Response, error) {
	uri := fmt.Sprintf("%s%s", c.Endpoint, req.URL)
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}
	req.URL = u

	req.Header.Set("Host", "cloud.culturedcode.com")
	req.Header.Set("Authorization", fmt.Sprintf("Password %s", c.password))
	req.Header.Set("User-Agent", ThingsUserAgent)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Encoding", "UTF8")
	req.Header.Set("Accept-Language", "en-us")

	return c.client.Do(req)
}
