// Copyright 2021 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package globodns

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
)

type Client struct {
	sync.Mutex

	client    *http.Client
	baseURL   string
	token     string
	userAgent string

	Bind   BindService
	Domain DomainService
	Record RecordService
}

func New(cli *http.Client, url string) (*Client, error) {
	if cli == nil {
		cli = http.DefaultClient
	}

	if url == "" {
		return nil, fmt.Errorf("globodns: URL cannot be empty")
	}

	c := &Client{
		client:    cli,
		baseURL:   strings.TrimSuffix(url, "/"),
		userAgent: "go-globodnsclient",
	}

	c.Bind = &bindService{Client: c}
	c.Domain = &domainService{Client: c}
	c.Record = &recordService{Client: c}

	return c, nil
}

func (c *Client) SetUserAgent(ua string) {
	c.userAgent = ua
}

func (c *Client) SetToken(token string) {
	c.Lock()
	defer c.Unlock()
	c.token = token
}

func (c *Client) Do(req *http.Request, out interface{}) (*http.Response, error) {
	if req == nil {
		return nil, fmt.Errorf("globodns: HTTP request cannot be nil")
	}

	if ua := req.Header.Get("User-Agent"); ua == "" {
		req.Header.Set("User-Agent", c.userAgent)
	}

	req.Header.Set("Accept", "application/json")

	if req.Method != "GET" && req.Method != "HEAD" {
		req.Header.Set("Content-Type", "application/json")
	}

	if c.token != "" {
		req.Header.Set("X-Auth-Token", c.token)
	}

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if err = checkResponse(res); err != nil {
		return res, err
	}

	if out == nil {
		return res, nil
	}

	if err = json.NewDecoder(res.Body).Decode(out); err != nil {
		return res, fmt.Errorf("globodns: failed to decode JSON object: %w", err)
	}

	return res, nil
}

func (c *Client) makeURL(path string) string {
	return fmt.Sprintf("%s/%s", c.baseURL, strings.TrimPrefix(path, "/"))
}

func checkResponse(res *http.Response) error {
	if res == nil {
		return fmt.Errorf("globodns: HTTP response cannot be nil")
	}

	if c := res.StatusCode; c >= 200 && c < 300 {
		return nil
	}

	return newUnexpectedHTTPError(res)
}

func newUnexpectedHTTPError(res *http.Response) error {
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("globodns: could not read the body message")
	}

	return fmt.Errorf("globodns: unexpected HTTP status code: Code: %d Body: %s", res.StatusCode, body)
}
