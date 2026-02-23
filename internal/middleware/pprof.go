package middleware

import (
	"net/http"
	"net/http/pprof"
	"runtime"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type MetricsCache struct {
	mu         sync.RWMutex
	lastUpdate time.Time
	ttl        time.Duration
	metrics    *RuntimeMetrics
}

type RuntimeMetrics struct {
	GoroutineCount int       `json:"goroutines"`
	MemoryAlloc    float64   `json:"memory_alloc_mb"`
	MemoryTotal    float64   `json:"memory_total_mb"`
	MemorySys      float64   `json:"memory_sys_mb"`
	HeapAlloc      float64   `json:"heap_alloc_mb"`
	HeapSys        float64   `json:"heap_sys_mb"`
	GCCount        uint32    `json:"gc_count"`
	GCPauseTotal   float64   `json:"gc_pause_s"`
	CPUCores       int       `json:"cpu_cores"`
	GoVersion      string    `json:"go_version"`
	Timestamp      time.Time `json:"timestamp"`
}

var globalMetricsCache = &MetricsCache{
	ttl: 5 * time.Second,
}

func init() {
	go metricsCollector()
}

func metricsCollector() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		collectMetrics()
	}
}

func collectMetrics() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	metrics := &RuntimeMetrics{
		GoroutineCount: runtime.NumGoroutine(),
		MemoryAlloc:    float64(m.Alloc) / 1024 / 1024,
		MemoryTotal:    float64(m.TotalAlloc) / 1024 / 1024,
		MemorySys:      float64(m.Sys) / 1024 / 1024,
		HeapAlloc:      float64(m.HeapAlloc) / 1024 / 1024,
		HeapSys:        float64(m.HeapSys) / 1024 / 1024,
		GCCount:        m.NumGC,
		GCPauseTotal:   float64(m.PauseTotalNs) / 1e9,
		CPUCores:       runtime.NumCPU(),
		GoVersion:      runtime.Version(),
		Timestamp:      time.Now(),
	}

	globalMetricsCache.mu.Lock()
	globalMetricsCache.metrics = metrics
	globalMetricsCache.lastUpdate = time.Now()
	globalMetricsCache.mu.Unlock()
}

func GetCachedMetrics() *RuntimeMetrics {
	globalMetricsCache.mu.RLock()
	defer globalMetricsCache.mu.RUnlock()

	if globalMetricsCache.metrics == nil {
		return &RuntimeMetrics{
			GoroutineCount: runtime.NumGoroutine(),
			CPUCores:       runtime.NumCPU(),
			GoVersion:      runtime.Version(),
			Timestamp:      time.Now(),
		}
	}

	return globalMetricsCache.metrics
}

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

func MetricsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		metrics := GetCachedMetrics()

		c.JSON(http.StatusOK, gin.H{
			"goroutines": metrics.GoroutineCount,
			"memory": gin.H{
				"alloc_mb":   metrics.MemoryAlloc,
				"total_mb":   metrics.MemoryTotal,
				"sys_mb":     metrics.MemorySys,
				"heap_alloc": metrics.HeapAlloc,
				"heap_sys":   metrics.HeapSys,
				"gc_count":   metrics.GCCount,
				"gc_pause_s": metrics.GCPauseTotal,
			},
			"cpu_cores":  metrics.CPUCores,
			"go_version": metrics.GoVersion,
			"timestamp":  metrics.Timestamp.Format(time.RFC3339),
		})
	}
}

func RegisterMetrics(r *gin.Engine) {
	r.GET("/debug/metrics", MetricsHandler())
	RegisterPProf(r)
}
