/*
 * Copyright 2019 Li Kexian
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * DNS over HTTPS (DoH) Golang Implementation
 * https://www.likexian.com/
 */

package cloudflare

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/likexian/doh-go"
	"github.com/likexian/gokit/xhttp"
	"github.com/likexian/gokit/xip"
	"strings"
)

// Client is a DoH provider client
type Client struct {
	provides int
	xhttp    *xhttp.Request
}

const (
	// DefaultProvides is default provides
	DefaultProvides = iota
)

var (
	// Upstream is DoH query upstream
	Upstream = map[int]string{
		DefaultProvides: "https://cloudflare-dns.com/dns-query",
	}
)

// Version returns package version
func Version() string {
	return "0.1.0"
}

// Author returns package author
func Author() string {
	return "[Li Kexian](https://www.likexian.com/)"
}

// License returns package license
func License() string {
	return "Licensed under the Apache License 2.0"
}

// New returns a new cloudflare provider client
func New(s ...string) *Client {
	return &Client{
		provides: DefaultProvides,
		xhttp:    xhttp.New(),
	}
}

// String returns string of provider
func (c *Client) String() string {
	return "cloudflare"
}

// SetProvides set upstream provides type, cloudflare does NOT supported
func (c *Client) SetProvides(p int) {
	c.provides = DefaultProvides
}

// Query do DoH query
func (c *Client) Query(ctx context.Context, d doh.Domain, t doh.Type) (*doh.Response, error) {
	return c.ECSQuery(ctx, d, t, "")
}

// ECSQuery do DoH query with the edns0-client-subnet option
func (c *Client) ECSQuery(ctx context.Context, d doh.Domain, t doh.Type, s doh.ECS) (*doh.Response, error) {
	param := xhttp.QueryParam{
		"name": strings.TrimSpace(string(d)),
		"type": strings.TrimSpace(string(t)),
	}

	ss := strings.TrimSpace(string(s))
	if ss != "" {
		ss, err := xip.FixSubnet(ss)
		if err != nil {
			return nil, err
		}
		param["edns_client_subnet"] = ss
	}

	rsp, err := c.xhttp.Get(Upstream[c.provides], param, ctx, xhttp.Header{"accept": "application/dns-json"})
	if err != nil {
		return nil, err
	}

	defer rsp.Close()
	buf, err := rsp.Bytes()
	if err != nil {
		return nil, err
	}

	rr := &doh.Response{}
	err = json.NewDecoder(bytes.NewBuffer(buf)).Decode(rr)
	if err != nil {
		return nil, err
	}

	if rr.Status != 0 {
		return rr, fmt.Errorf("doh: failed response code %d", rr.Status)
	}

	return rr, nil
}
