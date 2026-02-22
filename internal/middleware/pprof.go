package middleware

import (
	"net/http"
	"net/http/pprof"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
)

func PProfHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		switch c.Param("profile") {
		case "/":
			pprof.Index(c.Writer, c.Request)
		case "/cmdline":
			pprof.Cmdline(c.Writer, c.Request)
		case "/profile":
			pprof.Profile(c.Writer, c.Request)
		case "/symbol":
			pprof.Symbol(c.Writer, c.Request)
		case "/trace":
			pprof.Trace(c.Writer, c.Request)
		default:
			pprof.Handler(c.Param("profile")).ServeHTTP(c.Writer, c.Request)
		}
		c.Abort()
	}
}

func RegisterPProf(r *gin.Engine) {
	profileGroup := r.Group("/debug/pprof")
	{
		profileGroup.GET("/*profile", PProfHandler())
	}
}

type Metrics struct {
	GoroutineCount int
	MemoryAlloc    uint64
	MemoryTotal    uint64
	MemorySys      uint64
	GCCount        uint32
	GCPauseTotal   time.Duration
}

func GetRuntimeMetrics() Metrics {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return Metrics{
		GoroutineCount: runtime.NumGoroutine(),
		MemoryAlloc:    m.Alloc,
		MemoryTotal:    m.TotalAlloc,
		MemorySys:      m.Sys,
		GCCount:        m.NumGC,
		GCPauseTotal:   time.Duration(m.PauseTotalNs),
	}
}

func MetricsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		metrics := GetRuntimeMetrics()
		c.JSON(http.StatusOK, gin.H{
			"goroutines": metrics.GoroutineCount,
			"memory": gin.H{
				"alloc_mb":   float64(metrics.MemoryAlloc) / 1024 / 1024,
				"total_mb":   float64(metrics.MemoryTotal) / 1024 / 1024,
				"sys_mb":     float64(metrics.MemorySys) / 1024 / 1024,
				"gc_count":   metrics.GCCount,
				"gc_pause_s": metrics.GCPauseTotal.Seconds(),
			},
			"cpu_cores":  runtime.NumCPU(),
			"go_version": runtime.Version(),
		})
	}
}

func RegisterMetrics(r *gin.Engine) {
	r.GET("/debug/metrics", MetricsHandler())
	RegisterPProf(r)
}
