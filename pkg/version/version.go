// Copyright 2022 Greptime Team
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

package version

import (
	"fmt"
	"runtime"
)

var (
	gitCommit  = "none"
	gitVersion = "none"
	buildDate  = "none"
)

type Version struct {
	GitCommit  string
	GitVersion string
	GoVersion  string
	Compiler   string
	Platform   string
	BuildDate  string
}

func (v Version) String() string {
	format := "GitCommit: %s\n" +
		"GitVersion: %s\n" +
		"GoVersion: %s\n" +
		"Compiler: %s\n" +
		"Platform: %s\n" +
		"BuildDate: %s\n"
	return fmt.Sprintf(format, v.GitCommit, v.GitVersion, v.GoVersion, v.Compiler, v.Platform, v.BuildDate)
}

func Get() Version {
	return Version{
		GitCommit:  gitCommit,
		GitVersion: gitVersion,
		GoVersion:  runtime.Version(),
		Compiler:   runtime.Compiler,
		Platform:   fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		BuildDate:  buildDate,
	}
}
