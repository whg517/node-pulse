package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/whg517/node-pulse/pulse/internal/version"
)

// VersionHandler exposes build-time version metadata. Public (no auth):
// version/commit/build-date are not secrets and are useful for health checks,
// SRE triage, and the frontend's "about" panel. Closes the D-G5 gap from
// docs/user-journey.md §23.2.
//
// @Summary		Get server version
// @Description	Public build metadata: semantic version, git commit, build date, Go toolchain version.
// @Tags			meta
// @Produce		json
// @Success		200	{object}	version.Info
// @Router			/version [get]
func VersionHandler(c *gin.Context) {
	c.JSON(http.StatusOK, version.Get())
}
