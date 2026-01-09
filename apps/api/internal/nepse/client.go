// Package nepse provides a wrapper around go-nepse
package nepse

import (
	"os"

	"github.com/voidarchive/go-nepse"
)

type Client struct {
	api *nepse.Client
}

func NewClient() (*Client, error) {
	opts := nepse.DefaultOptions()
	opts.TLSVerification = os.Getenv("NEPSE_TLS_VERIFY") == "true"

	api, err := nepse.NewClient(opts)
	if err != nil {
		return nil, err
	}
	return &Client{api: api}, nil
}

func (c *Client) Close() error {
	return c.api.Close()
}
