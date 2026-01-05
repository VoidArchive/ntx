// Package nepse provides a wrapper arround go-nepse
package nepse

import "github.com/voidarchive/go-nepse"

type Client struct {
	api *nepse.Client
}

func NewClient() (*Client, error) {
	opts := nepse.DefaultOptions()
	opts.TLSVerification = false

	api, err := nepse.NewClient(opts)
	if err != nil {
		return nil, err
	}
	return &Client{api: api}, nil
}

func (c *Client) Close() error {
	return c.api.Close()
}
