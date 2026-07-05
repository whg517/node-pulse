// Package version exposes build-time metadata for the Pulse server.
//
// These variables are wired to ldflags in pulse/Makefile (build target) so a
// release build stamps the git tag/commit into the binary. They close the
// D-G5 gap from docs/user-journey.md §23.2 (previously ServiceVersion was
// hardcoded to "unknown" with no git-injected value).
//
// Example Makefile wiring (already present):
//
//	go build -ldflags "-X github.com/whg517/node-pulse/pulse/internal/version.Version=$(VERSION) \
//	                   -X github.com/whg517/node-pulse/pulse/internal/version.Commit=$(COMMIT) \
//	                   -X github.com/whg517/node-pulse/pulse/internal/version.BuildDate=$(BUILD_DATE)"
package version

import (
	"runtime"
	"runtime/debug"
)

// These are populated by -ldflags -X at build time; "dev" / "" defaults
// apply when running `go build` directly (e.g. during development).
var (
	// Version is the semantic version (e.g. "v1.2.3") or a git describe output.
	Version = "dev"
	// Commit is the full git SHA the binary was built from.
	Commit = ""
	// BuildDate is the UTC build timestamp in RFC3339.
	BuildDate = ""
)

// Info aggregates the build-time fields for JSON serialization.
type Info struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildDate string `json:"build_date"`
	GoVersion string `json:"go_version"`
}

// Get returns the current build info. When Version/Commit were not injected
// at build time (e.g. `go run`), it falls back to the embedded VCS info the
// toolchain stamps when building from a git worktree.
func Get() Info {
	v, commit, date := Version, Commit, BuildDate
	if v == "dev" || (commit == "" && v == "dev") {
		if bi, ok := debug.ReadBuildInfo(); ok {
			for _, setting := range bi.Settings {
				switch setting.Key {
				case "vcs.revision":
					if commit == "" {
						commit = setting.Value
					}
				case "vcs.time":
					if date == "" {
						date = setting.Value
					}
				case "vcs.modified":
					// If the worktree was dirty, flag the version.
					if setting.Value == "true" && v == "dev" {
						v = "dev-dirty"
					}
				}
			}
		}
	}
	return Info{
		Version:   v,
		Commit:    commit,
		BuildDate: date,
		GoVersion: runtime.Version(),
	}
}
