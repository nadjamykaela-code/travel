package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	skyscannerRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "skyscanner_requests_total",
			Help: "Total number of Skyscanner API requests",
		},
		[]string{"status"},
	)

	workerRunsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "worker_runs_total",
			Help: "Total number of worker runs",
		},
		[]string{"status"},
	)

	workerRunDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "worker_run_duration_seconds",
			Help:    "Worker run duration in seconds",
			Buckets: []float64{1, 5, 10, 30, 60, 120, 300, 600},
		},
		[]string{},
	)
)

func init() {
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
	prometheus.MustRegister(skyscannerRequestsTotal)
	prometheus.MustRegister(workerRunsTotal)
	prometheus.MustRegister(workerRunDuration)
}

func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		c.Next()

		status := strconv.Itoa(c.Writer.Status())
		httpRequestsTotal.WithLabelValues(c.Request.Method, path, status).Inc()
		httpRequestDuration.WithLabelValues(c.Request.Method, path).Observe(time.Since(start).Seconds())
	}
}

func PrometheusHandler() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

func RecordSkyscannerRequest(status string) {
	skyscannerRequestsTotal.WithLabelValues(status).Inc()
}

func RecordWorkerRun(status string, duration time.Duration) {
	workerRunsTotal.WithLabelValues(status).Inc()
	workerRunDuration.WithLabelValues().Observe(duration.Seconds())
}
