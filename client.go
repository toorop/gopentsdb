package gopentsdb

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

// ClientConfig represents the client configuration
// Endpoint: URL as string
// Username and Password: used for HTTP AUTH (optionals)
// Timeout: optional, defualt value is 0 -> system timeout
// InsecureSkipVerify: controls whether a client verifies the
// server's certificate chain and host name.
type ClientConfig struct {
	Endpoint           string
	Username           string
	Password           string
	Timeout            int
	InsecureSkipVerify bool
}

// Client represents an OpenSTDB client
type Client struct {
	endpoint   *url.URL
	username   string
	password   string
	httpClient *http.Client
}

// NewClient returns a new OpenSTDB Client
func NewClient(config ClientConfig) (client *Client, err error) {
	client = &Client{
		username:   config.Username,
		password:   config.Password,
		httpClient: new(http.Client),
	}
	if client.endpoint, err = url.Parse(config.Endpoint); err != nil {
		return nil, err
	}

	if config.Timeout != 0 {
		client.httpClient.Timeout = time.Duration(config.Timeout) * time.Second
	}

	if config.InsecureSkipVerify {
		client.httpClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}
	return
}

// Push pushes a slice of points to OpenSTDB
func (c *Client) Push(points []Point) error {
	JSONPoints, err := json.Marshal(points)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", c.endpoint.String()+"/api/put", bytes.NewBuffer(JSONPoints))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "")
	if c.username != "" {
		req.SetBasicAuth(c.username, c.password)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf(string(body))
	}
	return nil
}
