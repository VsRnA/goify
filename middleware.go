package goify

import (
	"fmt"
	"log"
	"strings"
	"time"
)

type MiddlewareFunc func(*Context, func())

func (rt *Router) Use(middleware ...MiddlewareFunc) {
	rt.middleware = append(rt.middleware, middleware...)
}

func (rt *Router) executeMiddleware(ctx *Context, handler HandlerFunc) {
	index := 0
	
	var next func()
	next = func() {
		if index < len(rt.middleware) {
			middleware := rt.middleware[index]
			index++
			middleware(ctx, next)
		} else {
			handler(ctx)
		}
	}
	
	next()
}

func Logger() MiddlewareFunc {
	return func(c *Context, next func()) {
		start := time.Now()
		method := c.Request.Method
		path := c.Request.URL.Path
		
		next()
		
		duration := time.Since(start)
		log.Printf("%s %s - %v", method, path, duration)
	}
}

func CORS() MiddlewareFunc {
	return CORSWithConfig(CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders: []string{"Content-Type", "Authorization"},
	})
}

type CORSConfig struct {
	AllowOrigins []string
	AllowMethods []string
	AllowHeaders []string
}

func CORSWithConfig(config CORSConfig) MiddlewareFunc {
	return func(c *Context, next func()) {
		origin := c.GetHeader("Origin")

		if len(config.AllowOrigins) == 0 {
			config.AllowOrigins = []string{"*"}
		}

		allowed := false
		for _, allowedOrigin := range config.AllowOrigins {
			if allowedOrigin == "*" || allowedOrigin == origin {
				allowed = true
				break
			}
		}
		
		if allowed {
			if origin != "" {
				c.SetHeader("Access-Control-Allow-Origin", origin)
			} else if len(config.AllowOrigins) == 1 && config.AllowOrigins[0] == "*" {
				c.SetHeader("Access-Control-Allow-Origin", "*")
			}
		}

		if len(config.AllowMethods) > 0 {
			methods := ""
			for i, method := range config.AllowMethods {
				if i > 0 {
					methods += ", "
				}
				methods += method
			}
			c.SetHeader("Access-Control-Allow-Methods", methods)
		}
		
		if len(config.AllowHeaders) > 0 {
			headers := ""
			for i, header := range config.AllowHeaders {
				if i > 0 {
					headers += ", "
				}
				headers += header
			}
			c.SetHeader("Access-Control-Allow-Headers", headers)
		}

		if c.Request.Method == "OPTIONS" {
			c.Response.WriteHeader(204)
			return
		}
		
		next()
	}
}

func Recovery() MiddlewareFunc {
	return func(c *Context, next func()) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v", err)
				c.SendInternalError("Internal server error")
			}
		}()
		
		next()
	}
}

func BasicAuth(username, password string) MiddlewareFunc {
	return func(c *Context, next func()) {
		user, pass, ok := c.Request.BasicAuth()
		if !ok || user != username || pass != password {
			c.SetHeader("WWW-Authenticate", `Basic realm="Restricted"`)
			c.SendUnauthorized("Authentication required")
			return
		}
		
		next()
	}
}

type RateLimiter struct {
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

func (rl *RateLimiter) Middleware() MiddlewareFunc {
	return func(c *Context, next func()) {
		ip := c.Request.RemoteAddr
		now := time.Now()

		if requests, exists := rl.requests[ip]; exists {
			var validRequests []time.Time
			for _, reqTime := range requests {
				if now.Sub(reqTime) < rl.window {
					validRequests = append(validRequests, reqTime)
				}
			}
			rl.requests[ip] = validRequests
		}

		if len(rl.requests[ip]) >= rl.limit {
			c.SendError(429, "Too many requests", "Rate limit exceeded")
			return
		}

		rl.requests[ip] = append(rl.requests[ip], now)
		
		next()
	}
}

func Static(prefix, root string) MiddlewareFunc {
	return func(c *Context, next func()) {
		if c.Request.Method != "GET" && c.Request.Method != "HEAD" {
			next()
			return
		}
		
		path := c.Request.URL.Path
		if len(prefix) > 0 && !strings.HasPrefix(path, prefix) {
			next()
			return
		}

		if len(prefix) > 0 {
			path = strings.TrimPrefix(path, prefix)
		}

		fullPath := root + path
		c.SendFile(fullPath)
	}
}

func RequestID() MiddlewareFunc {
	return func(c *Context, next func()) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		
		c.SetHeader("X-Request-ID", requestID)
		c.Set("requestID", requestID)
		
		next()
	}
}

func generateRequestID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}