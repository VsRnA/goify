package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/VsRnA/goify"
)

type User struct {
	ID       int    `json:"id" validate:"min=1"`
	Name     string `json:"name" validate:"required,min=2,max=50,alpha"`
	Email    string `json:"email" validate:"required,email"`
	Age      int    `json:"age" validate:"required,min=18,max=120"`
	Username string `json:"username" validate:"required,min=3,max=20,alphanum"`
	Website  string `json:"website,omitempty" validate:"url"`
	Role     string `json:"role" validate:"required,oneof=admin user moderator"`
}

type Product struct {
	Name        string  `json:"name" validate:"required,min=1,max=100"`
	Description string  `json:"description" validate:"max=500"`
	Price       float64 `json:"price" validate:"required,min=0"`
	Category    string  `json:"category" validate:"required,oneof=electronics clothing books"`
	Tags        []string `json:"tags" validate:"max=5"`
	InStock     bool    `json:"in_stock"`
}

type QueryParams struct {
	Page     int    `query:"page" validate:"min=1,max=1000"`
	Limit    int    `query:"limit" validate:"min=1,max=100"`
	Sort     string `query:"sort" validate:"oneof=name email age created_at"`
	Order    string `query:"order" validate:"oneof=asc desc"`
	Category string `query:"category" validate:"oneof=all electronics clothing books"`
}

func main() {
	app := goify.New()

	app.Use(goify.Logger())
	app.Use(goify.Recovery())
	app.Use(goify.CORS())

	goify.RegisterValidator("strong_password", func(value interface{}, param string) error {
		password, ok := value.(string)
		if !ok {
			return fmt.Errorf("strong_password validation only works on strings")
		}
		
		if len(password) < 8 {
			return fmt.Errorf("password must be at least 8 characters long")
		}
		
		hasUpper := false
		hasLower := false
		hasDigit := false
		
		for _, char := range password {
			switch {
			case 'A' <= char && char <= 'Z':
				hasUpper = true
			case 'a' <= char && char <= 'z':
				hasLower = true
			case '0' <= char && char <= '9':
				hasDigit = true
			}
		}
		
		if !hasUpper || !hasLower || !hasDigit {
			return fmt.Errorf("password must contain at least one uppercase letter, one lowercase letter, and one digit")
		}
		
		return nil
	})

	app.GET("/", func(c *goify.Context) {
		c.SendSuccess(goify.H{
			"message": "Welcome to Goify Validation Example!",
			"version": "0.4.0",
			"features": []string{
				"Struct validation with tags",
				"Query parameter validation", 
				"Custom validators",
				"Detailed error messages",
			},
		})
	})

	app.POST("/users", func(c *goify.Context) {
		var user User

		if err := c.BindAndValidate(&user); err != nil {
			c.SendValidationError(err)
			return
		}

		user.ID = 123
		
		c.SendCreated(user, "User created successfully")
	})

	app.PUT("/users/:id", func(c *goify.Context) {
		userID := c.Param("id")
		
		var user User
		if err := c.BindJSON(&user); err != nil {
			c.SendBadRequest("Invalid JSON format")
			return
		}

		if validationErrors := c.ValidateStruct(&user); len(validationErrors) > 0 {
			c.SendValidationError(validationErrors)
			return
		}

		user.ID, _ = strconv.Atoi(userID)
		c.SendSuccess(user, "User updated successfully")
	})

	app.POST("/products", func(c *goify.Context) {
		var product Product
		
		if err := c.BindAndValidate(&product); err != nil {
			c.SendValidationError(err)
			return
		}

		c.SendCreated(product, "Product created successfully")
	})

	app.GET("/users", func(c *goify.Context) {
		var params QueryParams

		params.Page = 1
		params.Limit = 10
		params.Sort = "name"
		params.Order = "asc"
		params.Category = "all"

		if err := c.ValidateQuery(&params); err != nil {
			c.SendValidationError(err)
			return
		}

		users := []goify.H{
			{"id": 1, "name": "Alice", "email": "alice@example.com", "age": 25},
			{"id": 2, "name": "Bob", "email": "bob@example.com", "age": 30},
		}

		c.SendSuccess(goify.H{
			"users": users,
			"pagination": goify.H{
				"page":  params.Page,
				"limit": params.Limit,
				"total": len(users),
			},
			"sort": goify.H{
				"field": params.Sort,
				"order": params.Order,
			},
			"filter": goify.H{
				"category": params.Category,
			},
		})
	})

	app.POST("/register", func(c *goify.Context) {
		var registerData struct {
			Username string `json:"username" validate:"required,min=3,max=20,alphanum"`
			Email    string `json:"email" validate:"required,email"`
			Password string `json:"password" validate:"required,strong_password"`
			Age      int    `json:"age" validate:"required,min=13"`
		}

		if err := c.BindAndValidate(&registerData); err != nil {
			c.SendValidationError(err)
			return
		}

		c.SendSuccess(goify.H{
			"message":  "Registration successful",
			"username": registerData.Username,
			"email":    registerData.Email,
		})
	})

	app.POST("/validate-manual", func(c *goify.Context) {
		var data map[string]interface{}
		
		if err := c.BindJSON(&data); err != nil {
			c.SendBadRequest("Invalid JSON")
			return
		}

		email, exists := data["email"].(string)
		if !exists || email == "" {
			c.SendFieldError("email", "Email is required")
			return
		}

		if !strings.Contains(email, "@") {
			c.SendFieldError("email", "Invalid email format")
			return
		}

		c.SendSuccess(goify.H{
			"message": "Manual validation passed",
			"data":    data,
		})
	})

	app.POST("/orders", func(c *goify.Context) {
		type Address struct {
			Street  string `json:"street" validate:"required,min=5"`
			City    string `json:"city" validate:"required,min=2"`
			Country string `json:"country" validate:"required,min=2"`
			Zip     string `json:"zip" validate:"required,numeric,min=5,max=10"`
		}

		type Order struct {
			CustomerName    string  `json:"customer_name" validate:"required,min=2"`
			Email          string  `json:"email" validate:"required,email"`
			Total          float64 `json:"total" validate:"required,min=0"`
			ShippingAddress Address `json:"shipping_address" validate:"required"`
			Items          []string `json:"items" validate:"required,min=1,max=10"`
		}

		var order Order
		
		if err := c.BindAndValidate(&order); err != nil {
			c.SendValidationError(err)
			return
		}

		c.SendCreated(goify.H{
			"order_id": 12345,
			"message":  "Order created successfully",
			"order":    order,
		})
	})

	log.Println("Server started with Validation support!")
	log.Println("")
	log.Println("Available endpoints:")
	log.Println("  GET  / - Welcome message")
	log.Println("  POST /users - Create user with validation")
	log.Println("  PUT  /users/:id - Update user")
	log.Println("  POST /products - Create product")
	log.Println("  GET  /users?page=1&limit=10 - Query validation")
	log.Println("  POST /register - Custom validator example")
	log.Println("  POST /validate-manual - Manual validation")
	log.Println("  POST /orders - Nested struct validation")
	log.Println("")
	log.Println("Try these example requests:")
	log.Println(`  curl -X POST http://localhost:3000/users \`)
	log.Println(`    -H "Content-Type: application/json" \`)
	log.Println(`    -d '{"name":"John","email":"john@example.com","age":25,"username":"john123","role":"admin"}'`)
	log.Println("")
	log.Println(`  curl "http://localhost:3000/users?page=0&limit=200&sort=invalid"`)
	
	if err := app.Listen(":3000"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}