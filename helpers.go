// Copyright 2021 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package globodns

import "time"

func BoolPointer(b bool) *bool { return &b }

func BoolValue(ptr *bool) bool {
	if ptr == nil {
		return false
	}

	return *ptr
}

func IntPointer(n int) *int { return &n }

func IntValue(ptr *int) int {
	if ptr == nil {
		return 0
	}

	return *ptr
}

func StringPointer(s string) *string { return &s }

func StringValue(ptr *string) string {
	if ptr == nil {
		return ""
	}

	return *ptr
}

func TimePointer(t time.Time) *time.Time { return &t }

func TimeValue(ptr *time.Time) time.Time {
	if ptr == nil {
		return time.Time{}
	}

	return *ptr
}

func TTL(r Record, d Domain) int {
	if ttl := r.GetTTL(); ttl != nil {
		return *ttl
	}

	if ttl := d.GetTTL(); ttl != nil {
		return *ttl
	}

	return 0
}
