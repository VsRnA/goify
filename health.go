package goify

import (
	"fmt"
	"runtime"
	"syscall"
	"time"
)

type HealthStatus string

const (
	StatusHealthy   HealthStatus = "healthy"
	StatusUnhealthy HealthStatus = "unhealthy"
	StatusDegraded  HealthStatus = "degraded"
)

type HealthCheck struct {
	Name        string
	Status      HealthStatus
	Message     string
	LastChecked time.Time
	Duration    time.Duration
	Data        interface{}
}

type HealthChecker func() HealthCheck

type HealthResponse struct {
	Status      HealthStatus           `json:"status"`
	Timestamp   time.Time              `json:"timestamp"`
	Uptime      string                 `json:"uptime"`
	Version     string                 `json:"version,omitempty"`
	Environment string                 `json:"environment,omitempty"`
	Checks      map[string]HealthCheck `json:"checks,omitempty"`
}

var (
	startTime     = time.Now()
	healthChecks  = make(map[string]HealthChecker)
	appVersion    = "1.0.0"
	appEnv        = "development"
)

func SetAppInfo(version, environment string) {
	appVersion = version
	appEnv = environment
}

func RegisterHealthCheck(name string, checker HealthChecker) {
	healthChecks[name] = checker
}

func HealthCheckMiddleware() MiddlewareFunc {
	return func(c *Context, next func()) {
		response := getHealthResponse()
		
		status := 200
		if response.Status != StatusHealthy {
			status = 503
		}
		
		c.JSON(status, response)
	}
}

func HealthCheckHandler() HandlerFunc {
	return func(c *Context) {
		response := getHealthResponse()
		
		status := 200
		if response.Status != StatusHealthy {
			status = 503
		}
		
		c.JSON(status, response)
	}
}

func getHealthResponse() HealthResponse {
	now := time.Now()
	uptime := now.Sub(startTime)
	
	checks := make(map[string]HealthCheck)
	overallStatus := StatusHealthy
	
	for name, checker := range healthChecks {
		start := time.Now()
		check := checker()
		check.Duration = time.Since(start)
		check.LastChecked = now
		
		checks[name] = check
		
		if check.Status == StatusUnhealthy {
			overallStatus = StatusUnhealthy
		} else if check.Status == StatusDegraded && overallStatus == StatusHealthy {
			overallStatus = StatusDegraded
		}
	}
	
	return HealthResponse{
		Status:      overallStatus,
		Timestamp:   now,
		Uptime:      formatDuration(uptime),
		Version:     appVersion,
		Environment: appEnv,
		Checks:      checks,
	}
}

func formatDuration(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60
	
	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm %ds", days, hours, minutes, seconds)
	} else if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}

func DatabaseHealthCheck(pingFunc func() error) HealthChecker {
	return func() HealthCheck {
		if err := pingFunc(); err != nil {
			return HealthCheck{
				Name:    "database",
				Status:  StatusUnhealthy,
				Message: err.Error(),
			}
		}
		
		return HealthCheck{
			Name:    "database",
			Status:  StatusHealthy,
			Message: "Database connection is healthy",
		}
	}
}

func RedisHealthCheck(pingFunc func() error) HealthChecker {
	return func() HealthCheck {
		if err := pingFunc(); err != nil {
			return HealthCheck{
				Name:    "redis",
				Status:  StatusUnhealthy,
				Message: err.Error(),
			}
		}
		
		return HealthCheck{
			Name:    "redis",
			Status:  StatusHealthy,
			Message: "Redis connection is healthy",
		}
	}
}

func MemoryHealthCheck(maxMemoryMB int) HealthChecker {
	return func() HealthCheck {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		
		memoryMB := int(m.Alloc / 1024 / 1024)
		
		if memoryMB > maxMemoryMB {
			return HealthCheck{
				Name:    "memory",
				Status:  StatusDegraded,
				Message: fmt.Sprintf("Memory usage is high: %dMB", memoryMB),
				Data: map[string]interface{}{
					"allocated_mb": memoryMB,
					"max_mb":       maxMemoryMB,
				},
			}
		}
		
		return HealthCheck{
			Name:    "memory",
			Status:  StatusHealthy,
			Message: fmt.Sprintf("Memory usage is normal: %dMB", memoryMB),
			Data: map[string]interface{}{
				"allocated_mb": memoryMB,
				"max_mb":       maxMemoryMB,
			},
		}
	}
}

func DiskSpaceHealthCheck(path string, minSpaceGB int) HealthChecker {
	return func() HealthCheck {
		var stat syscall.Statfs_t
		err := syscall.Statfs(path, &stat)
		if err != nil {
			return HealthCheck{
				Name:    "disk",
				Status:  StatusUnhealthy,
				Message: err.Error(),
			}
		}
		
		availableGB := int(stat.Bavail * uint64(stat.Bsize) / 1024 / 1024 / 1024)
		
		if availableGB < minSpaceGB {
			return HealthCheck{
				Name:    "disk",
				Status:  StatusDegraded,
				Message: fmt.Sprintf("Low disk space: %dGB available", availableGB),
				Data: map[string]interface{}{
					"available_gb": availableGB,
					"min_gb":       minSpaceGB,
					"path":         path,
				},
			}
		}
		
		return HealthCheck{
			Name:    "disk",
			Status:  StatusHealthy,
			Message: fmt.Sprintf("Disk space is sufficient: %dGB available", availableGB),
			Data: map[string]interface{}{
				"available_gb": availableGB,
				"min_gb":       minSpaceGB,
				"path":         path,
			},
		}
	}
}