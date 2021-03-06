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
		Bind:   &FakeBindService{},
		Domain: &FakeDomainService{},
		Record: &FakeRecordService{},
	}
}

var _ globodns.BindService = &FakeBindService{}

type FakeBindService struct {
	FakeExport func(ctx context.Context) (*globodns.ScheduleExport, error)
}

func (f *FakeBindService) Export(ctx context.Context) (*globodns.ScheduleExport, error) {
	if f.FakeExport != nil {
		return f.FakeExport(ctx)
	}

	return nil, nil
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
	FakeCreate func(ctx context.Context, r globodns.Record) (*globodns.Record, error)
	FakeDelete func(ctx context.Context, recordID int) error
	FakeList   func(ctx context.Context, domainID int, p *globodns.ListRecordsParameters) ([]globodns.Record, error)
	FakeUpdate func(ctx context.Context, r globodns.Record) error
}

func (f *FakeRecordService) Create(ctx context.Context, r globodns.Record) (*globodns.Record, error) {
	if f.FakeCreate == nil {
		return nil, fmt.Errorf("fake does not implement this method")
	}

	return f.FakeCreate(ctx, r)
}

func (f *FakeRecordService) Delete(ctx context.Context, recordID int) error {
	if f.FakeDelete == nil {
		return fmt.Errorf("fake does not implement this method")
	}

	return f.FakeDelete(ctx, recordID)
}

func (f *FakeRecordService) List(ctx context.Context, domainID int, p *globodns.ListRecordsParameters) ([]globodns.Record, error) {
	if f.FakeList == nil {
		return nil, fmt.Errorf("fake does not implement this method")
	}

	return f.FakeList(ctx, domainID, p)
}

func (f *FakeRecordService) Update(ctx context.Context, r globodns.Record) error {
	if f.FakeUpdate == nil {
		return fmt.Errorf("fake does not implement this method")
	}

	return f.FakeUpdate(ctx, r)
}
