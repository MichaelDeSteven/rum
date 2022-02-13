// Copyright 2017 Bo-Yi Wu.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

//go:build !jsoniter && !go_json
// +build !jsoniter,!go_json

package json

import "encoding/json"

var (
	// Marshal is exported by rum/json package.
	Marshal = json.Marshal
	// Unmarshal is exported by rum/json package.
	Unmarshal = json.Unmarshal
	// MarshalIndent is exported by rum/json package.
	MarshalIndent = json.MarshalIndent
	// NewDecoder is exported by rum/json package.
	NewDecoder = json.NewDecoder
	// NewEncoder is exported by rum/json package.
	NewEncoder = json.NewEncoder
)
