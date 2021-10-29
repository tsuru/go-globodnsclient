// Copyright 2021 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package globodns

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type BindService interface {
	Export(ctx context.Context) (*ScheduleExport, error)
}

type ScheduleExport struct {
	Output       string
	ScheduleDate time.Time
}

func (se *ScheduleExport) UnmarshalJSON(data []byte) error {
	var object map[string]interface{}
	if err := json.Unmarshal(data, &object); err != nil {
		return err
	}

	if output, ok := object["output"].(string); ok {
		se.Output = output
	}

	if dateStr, ok := object["schedule_date"].(string); ok {
		t, err := time.Parse("2006-01-02 15:04:05 -0700", dateStr)
		if err != nil {
			return err
		}

		se.ScheduleDate = t
	}

	return nil
}

var _ BindService = &bindService{}

func NewBindService(c *Client) BindService {
	return &bindService{Client: c}
}

type bindService struct {
	*Client
}

func (b *bindService) Export(ctx context.Context) (*ScheduleExport, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", b.makeURL("/bind9/schedule_export.json"), nil)
	if err != nil {
		return nil, err
	}

	var se ScheduleExport
	_, err = b.Do(req, &se)
	if err != nil {
		return nil, err
	}

	return &se, nil
}
