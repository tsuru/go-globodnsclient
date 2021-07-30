// Copyright 2021 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fake

import (
	"context"
	"fmt"

	globodns "github.com/tsuru/go-globodnsclient"
)

func New() *globodns.Client {
	return &globodns.Client{
		Domain: &FakeDomainService{},
		Record: &FakeRecordService{},
	}
}

var _ globodns.DomainService = &FakeDomainService{}

type FakeDomainService struct {
	FakeList func(ctx context.Context, p *globodns.ListDomainsParameters) ([]globodns.Domain, error)
}

func (f *FakeDomainService) List(ctx context.Context, p *globodns.ListDomainsParameters) ([]globodns.Domain, error) {
	if f.FakeList == nil {
		return nil, fmt.Errorf("fake does not implement this method")
	}

	return f.FakeList(ctx, p)
}

var _ globodns.RecordService = &FakeRecordService{}

type FakeRecordService struct {
	FakeList func(ctx context.Context, domainID int, p *globodns.ListRecordsParameters) ([]globodns.Record, error)
}

func (f *FakeRecordService) List(ctx context.Context, domainID int, p *globodns.ListRecordsParameters) ([]globodns.Record, error) {
	if f.FakeList == nil {
		return nil, fmt.Errorf("fake does not implement this method")
	}

	return f.FakeList(ctx, domainID, p)
}
