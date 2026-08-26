package observability

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// MetricsMiddleware records request totals, in-flight requests, and latency.
func MetricsMiddleware(m *Metrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		if m == nil {
			c.Next()
			return
		}
		m.IncInFlight()
		started := time.Now()
		defer func() {
			m.DecInFlight()
			model := c.GetString("model")
			if model == "" {
				model = c.GetString("requested_model")
			}
			if model == "" {
				model = c.GetString("ops_model")
			}
			if model == "" {
				model = "unknown"
			}
			status := c.Writer.Status()
			if status == 0 {
				status = http.StatusOK
			}
			m.IncRequests(strconv.Itoa(status), model)
			m.ObserveLatency(time.Since(started).Milliseconds(), status)
		}()
		c.Next()
	}
}
