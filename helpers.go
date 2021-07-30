// Copyright 2021 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package globodns

func BoolPointer(b bool) *bool { return &b }

func IntPointer(n int) *int { return &n }

func StringPointer(s string) *string { return &s }