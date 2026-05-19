// Package buildinfo exposes build-time metadata about the running binary.
//
// The package-level variables are overridden at link time via -ldflags:
//
//	go build -ldflags " \
//	  -X .../pkg/buildinfo.Version=v1.2.3 \
//	  -X .../pkg/buildinfo.Commit=$(git rev-parse HEAD) \
//	  -X .../pkg/buildinfo.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
//
// Defaults are what you get from `go run` (no ldflags), which keeps
// local development frictionless.
package buildinfo

var (
	Version   = "dev"
	Commit    = "unknown"
	BuildTime = "unknown"
)

// Snapshot is the JSON shape returned by /api/v1/version.
type Snapshot struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildTime string `json:"build_time"`
}

// Get returns the current build metadata.
func Get() Snapshot {
	return Snapshot{
		Version:   Version,
		Commit:    Commit,
		BuildTime: BuildTime,
	}
}
