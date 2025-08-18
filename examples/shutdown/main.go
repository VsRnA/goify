package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/VsRnA/goify"
)

func main() {
	app := goify.New()

	goify.SetAppInfo("1.0.0", "production")

	goify.RegisterHealthCheck("database", goify.DatabaseHealthCheck(func() error {
		if rand.Float32() < 0.1 {
			return fmt.Errorf("database connection timeout")
		}
		return nil
	}))

	goify.RegisterHealthCheck("memory", goify.MemoryHealthCheck(100))
	goify.RegisterHealthCheck("disk", goify.DiskSpaceHealthCheck("/", 1))

	app.Use(goify.Logger())
	app.Use(goify.Recovery())

	app.GET("/", func(c *goify.Context) {
		c.SendSuccess(goify.H{
			"message": "Сервер работает с graceful shutdown!",
			"time":    time.Now(),
		})
	})

	app.GET("/health", goify.HealthCheckHandler())

	app.GET("/liveness", func(c *goify.Context) {
		c.JSON(200, goify.H{
			"status": "alive",
			"time":   time.Now(),
		})
	})

	app.GET("/readiness", func(c *goify.Context) {
		response := goify.HealthResponse{
			Status:    goify.StatusHealthy,
			Timestamp: time.Now(),
		}
		c.JSON(200, response)
	})

	app.GET("/slow", func(c *goify.Context) {
		time.Sleep(5 * time.Second)
		c.SendSuccess(goify.H{
			"message": "Медленный ответ завершен",
		})
	})

	app.OnShutdown(func() {
		log.Println("Выполняется cleanup перед завершением...")
		time.Sleep(1 * time.Second)
		log.Println("Cleanup завершен")
	})

	app.OnShutdown(func() {
		log.Println("Закрытие соединений с базой данных...")
	})

	config := goify.ShutdownConfig{
		Timeout: 10 * time.Second,
	}

	log.Println("🚀 Сервер запущен с graceful shutdown поддержкой!")
	log.Println("Endpoints:")
	log.Println("  GET / - Основной endpoint")
	log.Println("  GET /health - Полная проверка здоровья")
	log.Println("  GET /liveness - Проверка живучести")
	log.Println("  GET /readiness - Проверка готовности")
	log.Println("  GET /slow - Медленный endpoint (5s)")
	log.Println("")
	log.Println("Для graceful shutdown отправьте SIGTERM или нажмите Ctrl+C")

	if err := app.ListenAndServeWithGracefulShutdown(":3000", config); err != nil {
		log.Printf("Ошибка при завершении сервера: %v", err)
	}

	log.Println("Сервер успешно завершен")
}