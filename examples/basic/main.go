package main

import (
	"log"
	"time"

	"github.com/VsRnA/goify"
)

func main() {
	app := goify.New()

	app.Use(goify.Logger())
	app.Use(goify.Recovery())
	app.Use(goify.CORS())
	app.Use(goify.RequestID())

	rateLimiter := goify.NewRateLimiter(10, time.Minute)
	app.Use(rateLimiter.Middleware())

	app.Use(func(c *goify.Context, next func()) {
		log.Println("Before request processing")
		start := time.Now()

		c.Set("start_time", start)
		
		next()
		
		duration := time.Since(start)
		log.Printf("Request completed in %v", duration)
	})

	app.GET("/", func(c *goify.Context) {
		requestID, _ := c.Get("requestID")
		
		c.SendSuccess(goify.H{
			"message":    "Hello from Goify with Middleware!",
			"request_id": requestID,
		})
	})

	app.GET("/panic", func(c *goify.Context) {
		panic("This is a test panic!")
	})

	app.Use(goify.BasicAuth("admin", "secret"))
	app.GET("/admin", func(c *goify.Context) {
		c.SendSuccess(goify.H{
			"message": "Welcome to admin panel!",
		})
	})

	app.Use(goify.Static("/static", "./static"))

	app.Use(goify.CORSWithConfig(goify.CORSConfig{
		AllowOrigins: []string{"http://localhost:3000", "https://myapp.com"},
		AllowMethods: []string{"GET", "POST"},
		AllowHeaders: []string{"Content-Type", "Authorization", "X-Custom-Header"},
	}))

	app.POST("/api/data", func(c *goify.Context) {
		var data map[string]interface{}
		if err := c.BindJSON(&data); err != nil {
			c.SendBadRequest("Invalid JSON")
			return
		}

		c.SendSuccess(goify.H{
			"received": data,
			"message":  "Data processed successfully",
		})
	})

	app.GET("/conditional", func(c *goify.Context) {
		if userID, exists := c.Get("user_id"); exists {
			c.SendSuccess(goify.H{
				"message": "Authenticated user",
				"user_id": userID,
			})
		} else {
			c.SendSuccess(goify.H{
				"message": "Anonymous user",
			})
		}
	})

	log.Println("Starting server with middleware examples...")
	log.Println("Try these endpoints:")
	log.Println("	 GET  / - Basic route with middleware")
	log.Println("	 GET  /panic - Test recovery middleware")
	log.Println("  GET  /admin - Protected with basic auth (admin:secret)")
	log.Println("  POST /api/data - CORS enabled")
	log.Println("  GET  /static/* - Static files")
	
	if err := app.Listen(":3000"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}