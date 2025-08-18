package goify

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type ShutdownConfig struct {
	Timeout         time.Duration
	ShutdownSignals []os.Signal
	OnShutdown      []func()
}

var globalShutdownCallbacks []func()

func DefaultShutdownConfig() ShutdownConfig {
	return ShutdownConfig{
		Timeout:         30 * time.Second,
		ShutdownSignals: []os.Signal{syscall.SIGTERM, syscall.SIGINT},
		OnShutdown:      globalShutdownCallbacks,
	}
}

func (rt *Router) ListenAndServeWithGracefulShutdown(addr string, config ...ShutdownConfig) error {
	cfg := DefaultShutdownConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	go func() {
		if err := rt.Listen(addr); err != nil && err != http.ErrServerClosed {
			log.Printf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, cfg.ShutdownSignals...)

	<-quit
	log.Println("Shutting down server...")

	for _, fn := range cfg.OnShutdown {
		fn()
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	return rt.Shutdown(ctx)
}

func (rt *Router) OnShutdown(fn func()) {
	globalShutdownCallbacks = append(globalShutdownCallbacks, fn)
}