// Copyright 2021 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package globodns

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type Domain struct {
	Name           string  `json:"name"`
	AuthorityType  string  `json:"authority_type"`
	AddressingType string  `json:"addressing_type"`
	Notes          *string `json:"notes"`
	ID             int     `json:"id"`
	TTL            int     `json:"ttl"`
	ViewID         int     `json:"view_id"`
}

type ListDomainRequest struct {
	Query          string
	View           string
	DomainsPerPage int
	Page           int
	Reverse        *bool
}

func (r *ListDomainRequest) AsURLValues() url.Values {
	data := make(url.Values)

	if r.Page != 0 {
		data.Set("page", strconv.Itoa(r.Page))
	}

	if r.DomainsPerPage != 0 {
		data.Set("per_page", strconv.Itoa(r.DomainsPerPage))
	}

	if r.Query != "" {
		data.Set("query", r.Query)
	}

	if r.Reverse != nil {
		data.Set("reverse", strconv.FormatBool(*r.Reverse))
	}

	if r.View != "" {
		data.Set("view", r.View)
	}

	return data
}

type DomainService interface {
	List(ctx context.Context, r ListDomainRequest) ([]Domain, error)
}

var _ DomainService = &domainService{}

func NewDomainService(c *Client) DomainService {
	return &domainService{Client: c}
}

type domainService struct {
	*Client
}

func (d *domainService) List(ctx context.Context, r ListDomainRequest) ([]Domain, error) {
	if r.Page < 0 {
		return nil, fmt.Errorf("globodns: page cannot be negative")
	}

	if r.DomainsPerPage < 0 {
		return nil, fmt.Errorf("globodns: domains per page cannot be negative")
	}

	if r.Page != 0 {
		return d.list(ctx, r)
	}

	return d.listAll(ctx, r)
}

func (d *domainService) listAll(ctx context.Context, r ListDomainRequest) ([]Domain, error) {
	var domains []Domain

	for page := 1; ; page++ {
		r.Page = page

		ds, err := d.list(ctx, r)
		if err != nil {
			return nil, err
		}

		if len(ds) == 0 {
			break
		}

		domains = append(domains, ds...)
	}

	return domains, nil
}

func (d *domainService) list(ctx context.Context, r ListDomainRequest) ([]Domain, error) {
	path := fmt.Sprintf("/domains?%s", r.AsURLValues().Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", d.makeURL(path), nil)
	if err != nil {
		return nil, err
	}

	res, err := d.Do(req)
	if err != nil {
		return nil, err
	}

	var ss []struct {
		Domain Domain `json:"domain"`
	}

	err = json.NewDecoder(res.Body).Decode(&ss)
	if err != nil {
		return nil, err
	}

	var domains []Domain
	for _, s := range ss {
		domains = append(domains, s.Domain)
	}

	return domains, nil
}
