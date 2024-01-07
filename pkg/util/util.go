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

// TODO(zyy17): Use generic to implement the following functions.

func StringPtr(s string) *string {
	if len(s) > 0 {
		return &s
	}
	return nil
}

func BoolPtr(b bool) *bool {
	return &b
}

func Uint64Ptr(i uint64) *uint64 {
	return &i
}
