#!/usr/bin/env bash

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
