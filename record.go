// Copyright 2021 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package globodns

import (
	"bytes"
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
	Content   string     `json:"content"`
	Name      string     `json:"name"`
	Type      string     `json:"type"`
	TTL       *string    `json:"ttl,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
	ID        int        `json:"id,omitempty"`
	DomainID  int        `json:"domain_id"`
}

func (r *Record) GetTTL() *int {
	if r == nil || r.TTL == nil {
		return nil
	}

	ttl, err := strconv.Atoi(*r.TTL)
	if err != nil {
		return nil
	}

	return &ttl
}

type RecordService interface {
	Create(ctx context.Context, r Record) (*Record, error)
	Delete(ctx context.Context, recordID int) error
	List(ctx context.Context, domainID int, p *ListRecordsParameters) ([]Record, error)
	Update(ctx context.Context, r Record) error
}

var _ RecordService = &recordService{}

type recordService struct {
	*Client
}

func (s *recordService) Create(ctx context.Context, r Record) (*Record, error) {
	if r.DomainID < 0 {
		return nil, fmt.Errorf("globodns: domain ID cannot be negative")
	}

	return s.create(ctx, r)
}

func (s *recordService) create(ctx context.Context, r Record) (*Record, error) {
	var body bytes.Buffer

	data := map[string]Record{"record": r}

	if err := json.NewEncoder(&body).Encode(&data); err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/domains/%d/records.json", r.DomainID)

	req, err := http.NewRequestWithContext(ctx, "POST", s.makeURL(path), &body)
	if err != nil {
		return nil, err
	}

	// NOTE: using anonymous struct to avoid problems while decoding JSON object
	// as it may contain unmapped fields, such as "warnings" one.
	var got struct {
		Record *Record `json:"record"`
	}

	_, err = s.Do(req, &got)
	if err != nil {
		return nil, err
	}

	if got.Record.Type == "" {
		got.Record.Type = strings.ToUpper(r.Type)
	}

	return got.Record, nil
}

func (s *recordService) Delete(ctx context.Context, recordID int) error {
	if recordID < 0 {
		return fmt.Errorf("globodns: record ID cannot be negative")
	}

	return s.delete(ctx, recordID)
}

func (s *recordService) delete(ctx context.Context, recordID int) error {
	path := fmt.Sprintf("/records/%d.json", recordID)

	req, err := http.NewRequestWithContext(ctx, "DELETE", s.makeURL(path), nil)
	if err != nil {
		return err
	}

	_, err = s.Do(req, nil)
	return err
}

type ListRecordsParameters struct {
	Reverse *bool
	Query   string
	PerPage int
	Page    int
}

func (p *ListRecordsParameters) Validate() error {
	if p == nil {
		return nil
	}

	if p.Page < 0 {
		return fmt.Errorf("globodns: page cannot be negative")
	}

	if p.PerPage < 0 {
		return fmt.Errorf("globodns: records per page cannot be negative")
	}

	return nil
}

func (p *ListRecordsParameters) AsURLValues() url.Values {
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

	return data
}

func (s *recordService) List(ctx context.Context, domainID int, p *ListRecordsParameters) ([]Record, error) {
	if domainID < 0 {
		return nil, fmt.Errorf("globodns: domain ID cannot be negative")
	}

	if err := p.Validate(); err != nil {
		return nil, err
	}

	if p == nil {
		p = &ListRecordsParameters{}
	}

	if p.Page != 0 {
		return s.list(ctx, domainID, p)
	}

	return s.listAll(ctx, domainID, p)
}

func (s *recordService) listAll(ctx context.Context, domainID int, p *ListRecordsParameters) ([]Record, error) {
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

func (s *recordService) list(ctx context.Context, domainID int, p *ListRecordsParameters) ([]Record, error) {
	path := fmt.Sprintf("/domains/%d/records.json?%s", domainID, p.AsURLValues().Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", s.makeURL(path), nil)
	if err != nil {
		return nil, err
	}

	var got []map[string]Record

	_, err = s.Do(req, &got)
	if err != nil {
		return nil, err
	}

	var records []Record
	for _, r := range got {
		for rtype, record := range r {
			record.Type = strings.ToUpper(rtype)
			records = append(records, record)
		}
	}

	return records, nil
}

func (s *recordService) Update(ctx context.Context, r Record) error {
	if r.ID < 0 {
		return fmt.Errorf("globodns: record ID cannot be negative")
	}

	return s.update(ctx, r)
}

func (s *recordService) update(ctx context.Context, r Record) error {
	var body bytes.Buffer
	data := map[string]Record{"record": r}

	if err := json.NewEncoder(&body).Encode(&data); err != nil {
		return err
	}

	path := fmt.Sprintf("/records/%d.json", r.ID)

	req, err := http.NewRequestWithContext(ctx, "PUT", s.makeURL(path), &body)
	if err != nil {
		return err
	}

	_, err = s.Do(req, nil)
	return err
}
