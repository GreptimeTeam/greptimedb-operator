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

package metasrv

import (
	"io/ioutil"
	"reflect"
	"testing"
)

func TestMetasrvConfigFromRawData(t *testing.T) {
	input, err := ioutil.ReadFile("../testdata/metasrv.toml")
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := FromRawData(input)
	if err != nil {
		t.Fatal(err)
	}

	ouput, err := cfg.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(input, ouput) {
		t.Errorf("generated config is not equal to original config:\n, want: %s\n, got: %s\n",
			string(input), string(ouput))
	}
}
