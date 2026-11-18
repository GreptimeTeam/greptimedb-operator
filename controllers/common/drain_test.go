// Copyright 2026 Greptime Team
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

package common

import "testing"

func TestDatanodeNodeID(t *testing.T) {
	tests := []struct {
		name        string
		ordinal     int32
		groupID     int32
		startNodeID int32
		want        uint64
	}{
		{
			name:    "single datanode without start node id",
			ordinal: 2,
			groupID: -1,
			want:    2,
		},
		{
			name:        "single datanode with start node id",
			ordinal:     2,
			groupID:     -1,
			startNodeID: 10,
			want:        12,
		},
		{
			name:    "datanode group without start node id",
			ordinal: 3,
			groupID: 1,
			want:    uint64(1)<<32 | 3,
		},
		{
			name:        "datanode group with start node id",
			ordinal:     3,
			groupID:     1,
			startNodeID: 10,
			want:        uint64(1)<<32 | 3 + 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DatanodeNodeID(tt.ordinal, tt.groupID, tt.startNodeID); got != tt.want {
				t.Fatalf("DatanodeNodeID() = %d, want %d", got, tt.want)
			}
		})
	}
}
