package middleware

import (
	"net/http"
	"strings"
	"time"

	"isac-cran-system/pkg/errors"
	"isac-cran-system/pkg/response"

	"github.com/gin-gonic/gin"
)

var apiKey = "isac-cran-api-key-2024"

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "missing authorization header")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(c, "invalid authorization format")
			c.Abort()
			return
		}

		token := parts[1]
		if token != apiKey {
			response.Unauthorized(c, "invalid api key")
			c.Abort()
			return
		}

		c.Set("authenticated", true)
		c.Next()
	}
}

func OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && parts[0] == "Bearer" && parts[1] == apiKey {
				c.Set("authenticated", true)
			}
		}
		c.Next()
	}
}

type RateLimiter struct {
	requests map[string]*clientInfo
	limit    int
	window   time.Duration
}

type clientInfo struct {
	count     int
	startTime time.Time
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string]*clientInfo),
		limit:    limit,
		window:   window,
	}
}

func (rl *RateLimiter) RateLimit() gin.HandlerFunc {
	go rl.cleanup()

	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		info, exists := rl.requests[clientIP]
		now := time.Now()

		if !exists || now.Sub(info.startTime) > rl.window {
			rl.requests[clientIP] = &clientInfo{
				count:     1,
				startTime: now,
			}
			c.Next()
			return
		}

		if info.count >= rl.limit {
			response.ErrorWithCode(c, errors.CodeServiceUnavailable, "rate limit exceeded")
			c.Abort()
			return
		}

		info.count++
		c.Next()
	}
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		now := time.Now()
		for ip, info := range rl.requests {
			if now.Sub(info.startTime) > rl.window*2 {
				delete(rl.requests, ip)
			}
		}
	}
}

func RateLimit(limit int, window time.Duration) gin.HandlerFunc {
	limiter := NewRateLimiter(limit, window)
	return limiter.RateLimit()
}

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		method := c.Request.Method
		clientIP := c.ClientIP()

		if query != "" {
			path = path + "?" + query
		}

		if status >= 400 {
			isacLogger.Error("HTTP request",
				"status", status,
				"method", method,
				"path", path,
				"latency", latency.String(),
				"client_ip", clientIP,
			)
		} else {
			isacLogger.Info("HTTP request",
				"status", status,
				"method", method,
				"path", path,
				"latency", latency.String(),
				"client_ip", clientIP,
			)
		}
	}
}

type simpleLogger struct{}

var isacLogger = &simpleLogger{}

func (l *simpleLogger) Info(msg string, fields ...interface{}) {
}

func (l *simpleLogger) Error(msg string, fields ...interface{}) {
}

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				isacLogger.Error("Panic recovered", "error", err, "path", c.Request.URL.Path)
				response.InternalError(c, "internal server error")
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-Requested-With")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Content-Type")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

func generateRequestID() string {
	return time.Now().Format("20060102150405") + randomString(8)
}

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().Nanosecond()%len(charset)]
	}
	return string(b)
}
