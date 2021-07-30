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
	"strings"
	"time"
)

type Record struct {
	Content   string    `json:"content"`
	Name      string    `json:"name"`
	TTL       *int      `json:"ttl"`
	ID        int       `json:"id"`
	DomainID  int       `json:"domain_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Type      string    `json:"-"`
}

type RecordService interface {
	List(ctx context.Context, domainID int, p *ListRecordParameters) ([]Record, error)
}

var _ RecordService = &recordService{}

type recordService struct {
	*Client
}

type ListRecordParameters struct {
	Reverse *bool
	Query   string
	PerPage int
	Page    int
}

func (p *ListRecordParameters) Validate() error {
	if p.Page < 0 {
		return fmt.Errorf("globodns: page cannot be negative")
	}

	if p.PerPage < 0 {
		return fmt.Errorf("globodns: records per page cannot be negative")
	}

	return nil
}

func (p *ListRecordParameters) AsURLValues() url.Values {
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

	return data
}

func (s *recordService) List(ctx context.Context, domainID int, p *ListRecordParameters) ([]Record, error) {
	if domainID < 0 {
		return nil, fmt.Errorf("globodns: domain ID cannot be negative")
	}

	if p == nil {
		p = new(ListRecordParameters)
	}

	if err := p.Validate(); err != nil {
		return nil, err
	}

	if p.Page != 0 {
		return s.list(ctx, domainID, p)
	}

	return s.listAll(ctx, domainID, p)
}

func (s *recordService) listAll(ctx context.Context, domainID int, p *ListRecordParameters) ([]Record, error) {
	var records []Record

	for page := 1; ; page++ {
		p.Page = page

		rs, err := s.list(ctx, domainID, p)
		if err != nil {
			return nil, err
		}

		if len(rs) == 0 {
			break
		}

		records = append(records, rs...)
	}

	return records, nil
}

func (s *recordService) list(ctx context.Context, domainID int, p *ListRecordParameters) ([]Record, error) {
	path := fmt.Sprintf("/domains/%d?%s", domainID, p.AsURLValues().Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", s.makeURL(path), nil)
	if err != nil {
		return nil, err
	}

	res, err := s.Do(req)
	if err != nil {
		return nil, err
	}

	var ss []map[string]Record

	err = json.NewDecoder(res.Body).Decode(&ss)
	if err != nil {
		return nil, err
	}

	var records []Record
	for _, r := range ss {
		for rtype, record := range r {
			record.Type = strings.ToUpper(rtype)
			records = append(records, record)
		}
	}

	return records, nil
}
