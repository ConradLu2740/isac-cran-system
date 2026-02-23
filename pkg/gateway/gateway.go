package gateway

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"isac-cran-system/pkg/discovery"

	"github.com/gin-gonic/gin"
)

type Route struct {
	Path      string
	Service   string
	Methods   []string
	StripPath bool
	RateLimit int
	Auth      bool
}

type Gateway struct {
	discovery   *discovery.ServiceDiscovery
	lb          discovery.LoadBalancer
	routes      []Route
	rateLimiter *RateLimiter
}

type RateLimiter struct {
	requests map[string]*ClientInfo
	limit    int
	window   time.Duration
}

type ClientInfo struct {
	Count     int
	ResetTime time.Time
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string]*ClientInfo),
		limit:    limit,
		window:   window,
	}
}

func (rl *RateLimiter) Allow(clientID string) bool {
	now := time.Now()
	info, exists := rl.requests[clientID]

	if !exists || now.After(info.ResetTime) {
		rl.requests[clientID] = &ClientInfo{
			Count:     1,
			ResetTime: now.Add(rl.window),
		}
		return true
	}

	if info.Count >= rl.limit {
		return false
	}

	info.Count++
	return true
}

func NewGateway(sd *discovery.ServiceDiscovery, lb discovery.LoadBalancer) *Gateway {
	return &Gateway{
		discovery:   sd,
		lb:          lb,
		rateLimiter: NewRateLimiter(100, time.Minute),
		routes:      make([]Route, 0),
	}
}

func (g *Gateway) AddRoute(route Route) {
	g.routes = append(g.routes, route)
}

func (g *Gateway) SetupRoutes(r *gin.Engine) {
	for _, route := range g.routes {
		handler := g.createProxyHandler(route)

		for _, method := range route.Methods {
			switch method {
			case "GET":
				r.GET(route.Path, handler)
			case "POST":
				r.POST(route.Path, handler)
			case "PUT":
				r.PUT(route.Path, handler)
			case "DELETE":
				r.DELETE(route.Path, handler)
			case "PATCH":
				r.PATCH(route.Path, handler)
			}
		}
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
}

func (g *Gateway) createProxyHandler(route Route) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientID := c.ClientIP()
		if !g.rateLimiter.Allow(clientID) {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}

		services, err := g.discovery.Discover(route.Service)
		if err != nil || len(services) == 0 {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "service unavailable"})
			return
		}

		service := g.lb.Select(services)
		if service == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "no available service"})
			return
		}

		targetURL, _ := url.Parse(fmt.Sprintf("http://%s:%d", service.Address, service.Port))
		proxy := httputil.NewSingleHostReverseProxy(targetURL)

		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			c.JSON(http.StatusBadGateway, gin.H{"error": "proxy error"})
		}

		if route.StripPath {
			c.Request.URL.Path = strings.TrimPrefix(c.Request.URL.Path, route.Path)
			if c.Request.URL.Path == "" {
				c.Request.URL.Path = "/"
			}
		}

		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

func (g *Gateway) Start(ctx context.Context, port int) error {
	r := gin.Default()

	g.SetupRoutes(r)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: r,
	}

	go func() {
		<-ctx.Done()
		srv.Shutdown(context.Background())
	}()

	return srv.ListenAndServe()
}

func DefaultRoutes() []Route {
	return []Route{
		{Path: "/api/v1/irs", Service: "irs-service", Methods: []string{"GET", "POST"}, StripPath: false},
		{Path: "/api/v1/algorithm", Service: "algorithm-service", Methods: []string{"GET", "POST"}, StripPath: false},
		{Path: "/api/v1/sensor", Service: "sensor-service", Methods: []string{"GET", "POST"}, StripPath: false},
		{Path: "/api/v1/channel", Service: "channel-service", Methods: []string{"GET", "POST"}, StripPath: false},
	}
}
