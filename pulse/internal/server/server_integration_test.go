package server

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/whg517/node-pulse/pulse/internal/config"
	"github.com/whg517/node-pulse/pulse/internal/db"
	"github.com/whg517/node-pulse/pulse/internal/testutil"
)

// TestServer_BuildAndShutdown exercises the full server lifecycle against a real
// test database: Build() wires routes + DB + scheduler + health checker, the
// health endpoint responds, and Shutdown() releases every resource.
//
// This is an integration test — it skips cleanly when no database is reachable
// (via db.SetupTestDB's testing.Short + ping guard) so `make test-unit` stays
// green.
func TestServer_BuildAndShutdown(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Bring up a migrated test database (skips without DB). SetupTestDB also
	// seeds the admin user; Build() dials the same DSN again.
	pool, cleanupDB := db.SetupTestDB(t)
	defer cleanupDB()

	// Sanity: confirm the schema was applied (nodes table exists).
	_, err := pool.Exec(context.Background(), "SELECT 1 FROM nodes LIMIT 1")
	require.NoError(t, err, "migrated schema must expose the nodes table")

	// Build a Config pointing at the same DSN Build() will dial.
	dsn := testutil.GetTestDBURL()
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port:         "0", // unused: we drive the router via httptest, not ListenAndServe
			ReadTimeout:  15,
			WriteTimeout: 15,
			IdleTimeout:  60,
			Mode:         "test",
		},
		DB: config.DatabaseConfig{
			URL:            dsn,
			MaxConnections: 5,
			MinConnections: 1,
		},
	}

	builder := &Builder{config: &Config{Config: cfg}}
	srv, err := builder.Build()
	require.NoError(t, err, "Build() should succeed with a live database")
	require.NotNil(t, srv)
	require.NotNil(t, srv.router, "router should be wired")
	require.NotNil(t, srv.database, "database should be initialised")
	require.NotNil(t, srv.scheduler, "scheduler should be initialised")
	require.NotNil(t, srv.healthChecker, "health checker should be initialised")
	require.NotNil(t, srv.cacheManager, "cache manager should be configured")

	// The routes are registered on srv.router. Drive the health endpoint via
	// httptest (no port binding) to prove route wiring end-to-end.
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	req.RequestURI = "/api/v1/health"
	srv.router.ServeHTTP(w, req)

	// Health may report degraded status (no scheduler running) but the route
	// must be reachable and return 5xx-or-2xx JSON, never 404.
	assert.NotEqual(t, http.StatusNotFound, w.Code, "health route must be registered")
	body, _ := io.ReadAll(w.Body)
	assert.NotEmpty(t, body, "health endpoint should return a body")

	// Graceful shutdown must not leak goroutines or error. Build() already set
	// up the shutdown context; just exercise Shutdown().
	err = srv.Shutdown()
	assert.NoError(t, err, "Shutdown() should release all resources cleanly")
}
