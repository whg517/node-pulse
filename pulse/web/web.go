// Package web embeds the compiled frontend (Vite production build) into the
// Pulse binary so a single binary serves both the API and the SPA.
//
// The embedded files live under dist/. A committed .gitkeep keeps the
// directory non-empty so //go:embed always compiles; run `make web-build` to
// populate dist with the real frontend before a release build.
package web

import (
	"embed"
	"io/fs"
)

//go:embed dist/*
var distFS embed.FS

// DistFS returns the embedded frontend filesystem rooted at the dist
// directory (so file paths are relative to dist/, e.g. "index.html",
// "assets/app.js"). Use it with http.FileServer or gin.StaticFS.
func DistFS() fs.FS {
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		// dist/ is embedded at compile time; fs.Sub can only fail if the path
		// is malformed, which is a programming error. Panic keeps the API simple.
		panic("web: invalid embedded dist path: " + err.Error())
	}
	return sub
}

// IndexHTML returns the SPA entry point HTML (index.html), used as the
// fallback for client-side routes. It returns an error when the frontend has
// not been built into dist; callers should treat that as "frontend unavailable".
func IndexHTML() ([]byte, error) {
	return distFS.ReadFile("dist/index.html")
}
