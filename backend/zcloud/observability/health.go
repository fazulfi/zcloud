package observability

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// RegisterHealthRoutes registers liveness and dependency readiness checks.
func RegisterHealthRoutes(r *gin.Engine, dbPinger func(ctx context.Context) error, redisPinger func(ctx context.Context) error, timeout time.Duration) {
	r.GET("/health/live", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/health/ready", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()
		checks := gin.H{"database": "ok", "redis": "ok"}
		failed := false
		if dbPinger == nil {
			checks["database"] = "error: database pinger unavailable"
			failed = true
		} else if err := dbPinger(ctx); err != nil {
			checks["database"] = fmt.Sprintf("error: %s", err)
			failed = true
		}
		if redisPinger == nil {
			checks["redis"] = "error: redis pinger unavailable"
			failed = true
		} else if err := redisPinger(ctx); err != nil {
			checks["redis"] = fmt.Sprintf("error: %s", err)
			failed = true
		}
		if failed {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "degraded", "checks": checks})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready", "checks": checks})
	})
}
