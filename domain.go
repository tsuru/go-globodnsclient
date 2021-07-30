// Copyright 2021 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package globodns

import (
	"context"
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
	TTL            *string `json:"ttl"`
	ID             int     `json:"id"`
	ViewID         int     `json:"view_id"`
}

func (d *Domain) GetTTL() *int {
	if d == nil || d.TTL == nil {
		return nil
	}

	ttl, err := strconv.Atoi(*d.TTL)
	if err != nil {
		return nil
	}

	return &ttl
}

type ListDomainsParameters struct {
	Query   string
	View    string
	Reverse *bool
	Page    int
	PerPage int
}

func (p *ListDomainsParameters) Validate() error {
	if p == nil {
		return nil
	}

	if p.Page < 0 {
		return fmt.Errorf("globodns: page cannot be negative")
	}

	if p.PerPage < 0 {
		return fmt.Errorf("globodns: domains per page cannot be negative")
	}

	return nil
}

func (p *ListDomainsParameters) AsURLValues() url.Values {
	if p == nil {
		return nil
	}

	data := make(url.Values)
	if p.Page != 0 {
		data.Set("page", strconv.Itoa(p.Page))
	}

	if p.PerPage != 0 {
		data.Set("per_page", strconv.Itoa(p.PerPage))
	}

	if p.Query != "" {
		data.Set("query", p.Query)
	}

	if p.Reverse != nil {
		data.Set("reverse", strconv.FormatBool(*p.Reverse))
	}

	if p.View != "" {
		data.Set("view", p.View)
	}

	return data
}

type DomainService interface {
	List(ctx context.Context, p *ListDomainsParameters) ([]Domain, error)
}

var _ DomainService = &domainService{}

func NewDomainService(c *Client) DomainService {
	return &domainService{Client: c}
}

type domainService struct {
	*Client
}

func (d *domainService) List(ctx context.Context, p *ListDomainsParameters) ([]Domain, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}

	if p == nil {
		p = &ListDomainsParameters{}
	}

	if p.Page != 0 {
		return d.list(ctx, p)
	}

	return d.listAll(ctx, p)
}

func (d *domainService) listAll(ctx context.Context, p *ListDomainsParameters) ([]Domain, error) {
	var domains []Domain

	for page := 1; ; page++ {
		p.Page = page

		ds, err := d.list(ctx, p)
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

func (d *domainService) list(ctx context.Context, p *ListDomainsParameters) ([]Domain, error) {
	path := fmt.Sprintf("/domains?%s", p.AsURLValues().Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", d.makeURL(path), nil)
	if err != nil {
		return nil, err
	}

	var got []struct {
		Domain Domain `json:"domain"`
	}

	_, err = d.Do(req, &got)
	if err != nil {
		return nil, err
	}

	var domains []Domain
	for _, s := range got {
		domains = append(domains, s.Domain)
	}

	return domains, nil
}
