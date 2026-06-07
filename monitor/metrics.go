package monitor

import (
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
		[]string{"method", "path"},
	)
	httpRequestDurationHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "Histogram of the duration of HTTP requests",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path"})
)

func init() {
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDurationHistogram)
}

func RegisterMetrics(engine *gin.Engine) {
	engine.Use(func(c *gin.Context) {
		if c.Request.URL.Path == "/metrics" {
			return
		}

		start := time.Now()
		c.Next()

		//记录请求次数
		httpRequestsTotal.WithLabelValues(c.Request.Method, c.Request.URL.Path).Inc()
		//记录http方法和路径对应的耗时
		httpRequestDurationHistogram.WithLabelValues(c.Request.Method, c.Request.URL.Path).Observe(time.Since(start).Seconds())
	})

	engine.GET("/metrics", gin.WrapH(promhttp.Handler()))
}
