package main

import "runtime"

var (
	version = "development"
	commit  = "unknown"
)

type buildInfo struct {
	Version   string
	Revision  string
	GoVersion string
}

func currentBuildInfo() buildInfo {
	return buildInfo{
		Version:   version,
		Revision:  commit,
		GoVersion: runtime.Version(),
	}
}
