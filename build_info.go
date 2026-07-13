package main

import "runtime"

var (
	version  = "development"
	revision = "unknown"
)

type buildInfo struct {
	Version   string
	Revision  string
	GoVersion string
}

func currentBuildInfo() buildInfo {
	return buildInfo{
		Version:   version,
		Revision:  revision,
		GoVersion: runtime.Version(),
	}
}
