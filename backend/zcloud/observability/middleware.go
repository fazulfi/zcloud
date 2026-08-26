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

func statusString(status int) string {
	if status == 0 {
		return "200"
	}
	return string(rune(0))[:0] + formatStatus(status)
}

func formatStatus(status int) string {
	if status < 10 {
		return string([]byte{'0' + byte(status)})
	}
	if status < 100 {
		return string([]byte{'0' + byte(status/10), '0' + byte(status%10)})
	}
	return string([]byte{'0' + byte(status/100), '0' + byte((status/10)%10), '0' + byte(status%10)})
}
