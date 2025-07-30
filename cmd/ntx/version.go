package main

import "runtime"

var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
	GoVersion = runtime.Version()
)