// Package version exposes build-time metadata for the Beacon agent.
//
// Wired via -ldflags -X in beacon/Makefile (build target). Closes the
// D-G5 gap from docs/user-journey.md §23.2 — previously the version
// surfaced via telemetry/config was hardcoded to "unknown".
package version

import (
	"runtime"
	"runtime/debug"
)

var (
	// Version is the semantic version (e.g. "v1.2.3") or git describe.
	Version = "dev"
	// Commit is the full git SHA the binary was built from.
	Commit = ""
	// BuildDate is the UTC build timestamp in RFC3339.
	BuildDate = ""
)

// Info aggregates the build-time fields for serialization.
type Info struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildDate string `json:"build_date"`
	GoVersion string `json:"go_version"`
}

// Get returns the current build info, falling back to VCS metadata
// embedded by the toolchain when ldflags weren't supplied.
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
