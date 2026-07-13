package main

import (
	"runtime"
	"testing"
)

func TestCurrentBuildInfo(t *testing.T) {
	build := currentBuildInfo()

	if build.Version != version {
		t.Fatalf("Version = %q, want %q", build.Version, version)
	}
	if build.Revision != commit {
		t.Fatalf("Revision = %q, want %q", build.Revision, commit)
	}
	if build.GoVersion != runtime.Version() {
		t.Fatalf("GoVersion = %q, want %q", build.GoVersion, runtime.Version())
	}
}
