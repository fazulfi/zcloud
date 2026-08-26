package observability

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
)

var latencyBuckets = [...]int64{5, 10, 25, 50, 100, 250, 500, 1000, 2500, 5000, 10000}

// Metrics stores lightweight HTTP request metrics without an external metrics client.
type Metrics struct {
	requests     atomic.Int64
	inFlight     atomic.Int64
	mu           sync.Mutex
	statuses     map[string]int64
	models       map[string]int64
	buckets      map[int]int64
	latencyCount atomic.Int64
	latencySum   atomic.Int64
}

// NewMetrics creates an empty metrics registry.
func NewMetrics() *Metrics {
	return &Metrics{
		statuses: make(map[string]int64),
		models:   make(map[string]int64),
		buckets:  make(map[int]int64),
	}
}

// IncRequests records a completed request by status and model.
func (m *Metrics) IncRequests(status, model string) {
	if m == nil {
		return
	}
	m.requests.Add(1)
	m.mu.Lock()
	m.statuses[status]++
	m.models[model]++
	m.mu.Unlock()
}

// IncInFlight increments the number of active requests.
func (m *Metrics) IncInFlight() {
	if m != nil {
		m.inFlight.Add(1)
	}
}

// DecInFlight decrements the number of active requests.
func (m *Metrics) DecInFlight() {
	if m != nil {
		m.inFlight.Add(-1)
	}
}

// ObserveLatency records a request duration in milliseconds and its status bucket.
func (m *Metrics) ObserveLatency(ms int64, _ int) {
	if m == nil {
		return
	}
	m.latencyCount.Add(1)
	m.latencySum.Add(ms)
	m.mu.Lock()
	for i, bound := range latencyBuckets {
		if ms <= bound {
			m.buckets[i]++
		}
	}
	m.mu.Unlock()
}

// Snapshot returns metrics in Prometheus text exposition format.
func (m *Metrics) Snapshot() string {
	if m == nil {
		return ""
	}
	m.mu.Lock()
	statuses := cloneMap(m.statuses)
	models := cloneMap(m.models)
	buckets := make(map[int]int64, len(m.buckets))
	for k, v := range m.buckets {
		buckets[k] = v
	}
	m.mu.Unlock()

	var b strings.Builder
	_, _ = b.WriteString("# HELP zcloud_http_requests_total Total HTTP requests.\n# TYPE zcloud_http_requests_total counter\n")
	keys := sortedKeys(statuses)
	for _, k := range keys {
		fmt.Fprintf(&b, "zcloud_http_requests_total{status=\"%s\"} %d\n", escape(k), statuses[k])
	}
	_, _ = b.WriteString("# HELP zcloud_http_requests_in_flight Current HTTP requests in flight.\n# TYPE zcloud_http_requests_in_flight gauge\n")
	fmt.Fprintf(&b, "zcloud_http_requests_in_flight %d\n", m.inFlight.Load())
	_, _ = b.WriteString("# HELP zcloud_http_requests_by_model_total Total HTTP requests by model.\n# TYPE zcloud_http_requests_by_model_total counter\n")
	for _, k := range sortedKeys(models) {
		fmt.Fprintf(&b, "zcloud_http_requests_by_model_total{model=\"%s\"} %d\n", escape(k), models[k])
	}
	_, _ = b.WriteString("# HELP zcloud_http_request_duration_ms Request duration in milliseconds.\n# TYPE zcloud_http_request_duration_ms histogram\n")
	for i, bound := range latencyBuckets {
		fmt.Fprintf(&b, "zcloud_http_request_duration_ms_bucket{le=\"%d\"} %d\n", bound, buckets[i])
	}
	fmt.Fprintf(&b, "zcloud_http_request_duration_ms_bucket{le=\"+Inf\"} %d\n", m.latencyCount.Load())
	fmt.Fprintf(&b, "zcloud_http_request_duration_ms_sum %d\nzcloud_http_request_duration_ms_count %d\n", m.latencySum.Load(), m.latencyCount.Load())
	return b.String()
}

func cloneMap(src map[string]int64) map[string]int64 {
	dst := make(map[string]int64, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
func sortedKeys(src map[string]int64) []string {
	keys := make([]string, 0, len(src))
	for k := range src {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
func escape(s string) string {
	return strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\n", "\\n").Replace(s)
}
