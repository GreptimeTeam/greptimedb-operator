// Copyright 2023 Greptime Team
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"crypto/sha256"
	"encoding/hex"
)

// MergeStringMap merges two string maps.
func MergeStringMap(origin, new map[string]string) map[string]string {
	if origin == nil {
		origin = make(map[string]string)
	}

	for k, v := range new {
		origin[k] = v
	}

	return origin
}

// CalculateConfigHash calculates the sha256 hash of the given config.
func CalculateConfigHash(config []byte) string {
	if len(config) == 0 {
		return ""
	}

	hash := sha256.Sum256(config)
	return hex.EncodeToString(hash[:])
}
