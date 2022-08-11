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
