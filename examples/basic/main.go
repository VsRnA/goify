package main

import (
	"log"

	"github.com/VsRnA/goify"
)

func main() {
	app := goify.New()

	app.GET("/", func(c *goify.Context) {
		c.JSON(200, goify.H{
			"message": "Hello from Goify!",
			"version": "0.1.0",
		})
	})

	app.GET("/hello", func(c *goify.Context) {
		name := c.QueryDefault("name", "World")
		c.SendSuccess(goify.H{
			"greeting": "Hello, " + name + "!",
		}, "Greeting sent successfully")
	})

	app.POST("/users", func(c *goify.Context) {
		var user struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		}

		if err := c.BindJSON(&user); err != nil {
			c.SendBadRequest("Invalid JSON format", err.Error())
			return
		}

		if user.Name == "" || user.Email == "" {
			c.SendBadRequest("Name and email are required")
			return
		}

		response := goify.H{
			"id":    123,
			"name":  user.Name,
			"email": user.Email,
		}

		c.SendCreated(response, "User created successfully")
	})

	app.GET("/users/profile", func(c *goify.Context) {
		userID := c.Query("id")
		if userID == "" {
			c.SendBadRequest("User ID is required")
			return
		}

		c.SendSuccess(goify.H{
			"id":   userID,
			"name": "John Doe",
			"email": "john@example.com",
		})
	})

	app.GET("/demo", func(c *goify.Context) {
		responseType := c.QueryDefault("type", "json")

		switch responseType {
		case "text":
			c.String(200, "Hello, this is plain text response!")
		case "html":
			c.HTML(200, "<h1>Hello HTML!</h1><p>This is an HTML response from Goify.</p>")
		case "error":
			c.SendNotFound("This is a demo 404 error")
		case "redirect":
			c.Redirect(302, "/")
		default:
			c.SendSuccess(goify.H{
				"message": "This is a JSON response",
				"type":    "demo",
			})
		}
	})

	app.GET("/headers", func(c *goify.Context) {
		userAgent := c.GetHeader("User-Agent")
		c.SetHeader("X-Custom-Header", "Goify-Framework")
		
		c.SendSuccess(goify.H{
			"your_user_agent": userAgent,
			"custom_header_set": "X-Custom-Header",
		})
	})

	log.Println("Starting Goify basic example...")
	if err := app.Listen(":3000"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}