#!/usr/bin/env bash
# Copyright 2022 Greptime Team
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


function ldflag() {
    local key=${1}
    local val=${2}

    echo "-X 'github.com/GreptimeTeam/greptimedb-operator/pkg/version.${key}=${val}'"
}

# parse the current git commit hash
GitCommit=$(git rev-parse HEAD)

# check if the current commit has a matching tag
GitVersion=$(git describe --exact-match --abbrev=0 --tags "${GitCommit}" 2> /dev/null || true)

# check for changed files (not untracked files)
if [ -n "${GitVersion}" ] && [ -n "$(git diff --shortstat 2> /dev/null | tail -n1)" ]; then
    GitVersion+="${GitVersion}-dirty"
fi

ldflags+=($(ldflag "gitCommit" "${GitCommit}"))
ldflags+=($(ldflag "gitVersion" "${GitVersion}"))
ldflags+=($(ldflag "buildDate" "$(date ${buildDate} -u +'%Y-%m-%dT%H:%M:%SZ')"))

echo "${ldflags[*]-}"
