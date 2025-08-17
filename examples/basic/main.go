package main

import (
	"log"
	"strconv"

	"github.com/VsRnA/goify"
)

func main() {
	app := goify.New()

	app.Use(goify.Logger())
	app.Use(goify.Recovery())
	app.Use(goify.CORS())

	app.GET("/", func(c *goify.Context) {
		c.SendSuccess(goify.H{
			"message": "Welcome to Goify with Groups and Parameters!",
			"version": "0.3.0",
		})
	})

	app.GET("/users/:id", func(c *goify.Context) {
		userID := c.Param("id")

		id, err := strconv.Atoi(userID)
		if err != nil {
			c.SendBadRequest("Invalid user ID")
			return
		}

		c.SendSuccess(goify.H{
			"user_id": id,
			"message": "User retrieved successfully",
			"user": goify.H{
				"id":   id,
				"name": "John Doe",
				"email": "john@example.com",
			},
		})
	})

	app.GET("/users/:userId/posts/:postId", func(c *goify.Context) {
		userID := c.Param("userId")
		postID := c.Param("postId")

		c.SendSuccess(goify.H{
			"user_id": userID,
			"post_id": postID,
			"message": "Post retrieved successfully",
			"post": goify.H{
				"id":      postID,
				"title":   "Sample Post",
				"author":  userID,
				"content": "This is a sample post content...",
			},
		})
	})

	app.GET("/files/*filepath", func(c *goify.Context) {
		filepath := c.Param("filepath")
		
		c.SendSuccess(goify.H{
			"filepath": filepath,
			"message":  "File path captured",
			"type":     "wildcard",
		})
	})

	v1 := app.Group("/api/v1")
	v1.Use(func(c *goify.Context, next func()) {
		c.SetHeader("API-Version", "v1")
		c.Set("api_version", "v1")
		next()
	})

	users := v1.Group("/users")
	users.GET("", func(c *goify.Context) {
		page := c.QueryDefault("page", "1")
		limit := c.QueryDefault("limit", "10")

		c.SendSuccess(goify.H{
			"users": []goify.H{
				{"id": 1, "name": "Alice", "email": "alice@example.com"},
				{"id": 2, "name": "Bob", "email": "bob@example.com"},
			},
			"pagination": goify.H{
				"page":  page,
				"limit": limit,
				"total": 2,
			},
		})
	})

	users.GET("/:id", func(c *goify.Context) {
		userID := c.Param("id")
		apiVersion, _ := c.Get("api_version")

		c.SendSuccess(goify.H{
			"user_id":     userID,
			"api_version": apiVersion,
			"user": goify.H{
				"id":    userID,
				"name":  "Alice Johnson",
				"email": "alice@example.com",
			},
		})
	})

	users.POST("", func(c *goify.Context) {
		var user struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		}

		if err := c.BindJSON(&user); err != nil {
			c.SendBadRequest("Invalid JSON format")
			return
		}

		if user.Name == "" || user.Email == "" {
			c.SendBadRequest("Name and email are required")
			return
		}

		c.SendCreated(goify.H{
			"id":    123,
			"name":  user.Name,
			"email": user.Email,
		}, "User created successfully")
	})

	users.PUT("/:id", func(c *goify.Context) {
		userID := c.Param("id")
		
		var user struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		}

		if err := c.BindJSON(&user); err != nil {
			c.SendBadRequest("Invalid JSON format")
			return
		}

		c.SendSuccess(goify.H{
			"id":      userID,
			"name":    user.Name,
			"email":   user.Email,
			"updated": true,
		}, "User updated successfully")
	})

	users.DELETE("/:id", func(c *goify.Context) {
		userID := c.Param("id")
		
		c.SendSuccess(goify.H{
			"id":      userID,
			"deleted": true,
		}, "User deleted successfully")
	})

	posts := users.Group("/:userId/posts")
	posts.GET("", func(c *goify.Context) {
		userID := c.Param("userId")
		
		c.SendSuccess(goify.H{
			"user_id": userID,
			"posts": []goify.H{
				{"id": 1, "title": "First Post", "author": userID},
				{"id": 2, "title": "Second Post", "author": userID},
			},
		})
	})

	posts.GET("/:postId", func(c *goify.Context) {
		userID := c.Param("userId")
		postID := c.Param("postId")
		
		c.SendSuccess(goify.H{
			"user_id": userID,
			"post_id": postID,
			"post": goify.H{
				"id":      postID,
				"title":   "Sample Post Title",
				"content": "This is the post content...",
				"author":  userID,
			},
		})
	})

	v2 := app.Group("/api/v2")
	v2.Use(func(c *goify.Context, next func()) {
		c.SetHeader("API-Version", "v2")
		next()
	})
	v2.Use(goify.BasicAuth("api", "secret"))

	v2.GET("/profile", func(c *goify.Context) {
		c.SendSuccess(goify.H{
			"message": "This is API v2 with authentication",
			"profile": goify.H{
				"username": "api",
				"role":     "admin",
			},
		})
	})

	admin := app.Group("/admin")
	admin.Use(goify.BasicAuth("admin", "supersecret"))
	admin.Use(func(c *goify.Context, next func()) {
		log.Println("Admin access logged")
		c.Set("role", "admin")
		next()
	})

	admin.GET("/dashboard", func(c *goify.Context) {
		role, _ := c.Get("role")
		
		c.SendSuccess(goify.H{
			"message": "Welcome to admin dashboard",
			"role":    role,
			"stats": goify.H{
				"users":    1250,
				"posts":    5670,
				"comments": 12340,
			},
		})
	})

	admin.GET("/users/:id/details", func(c *goify.Context) {
		userID := c.Param("id")
		
		c.SendSuccess(goify.H{
			"user_id": userID,
			"details": goify.H{
				"login_count":    45,
				"last_login":     "2024-01-15T10:30:00Z",
				"account_status": "active",
				"permissions":    []string{"read", "write", "delete"},
			},
		})
	})

	log.Println("🚀 Server started with Groups and Parameters support!")
	log.Println("")
	log.Println("Available endpoints:")
	log.Println("  GET  / - Welcome message")
	log.Println("  GET  /users/:id - Single parameter")
	log.Println("  GET  /users/:userId/posts/:postId - Multiple parameters")
	log.Println("  GET  /files/*filepath - Wildcard parameter")
	log.Println("")
	log.Println("API v1 Group (/api/v1):")
	log.Println("  GET  /api/v1/users - List users")
	log.Println("  GET  /api/v1/users/:id - Get user")
	log.Println("  POST /api/v1/users - Create user")
	log.Println("  PUT  /api/v1/users/:id - Update user")
	log.Println("  DELETE /api/v1/users/:id - Delete user")
	log.Println("  GET  /api/v1/users/:userId/posts - User posts")
	log.Println("  GET  /api/v1/users/:userId/posts/:postId - Specific post")
	log.Println("")
	log.Println("API v2 Group (/api/v2) - Auth: api:secret:")
	log.Println("  GET  /api/v2/profile - User profile")
	log.Println("")
	log.Println("Admin Group (/admin) - Auth: admin:supersecret:")
	log.Println("  GET  /admin/dashboard - Admin dashboard")
	log.Println("  GET  /admin/users/:id/details - User details")
	
	if err := app.Listen(":3000"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}