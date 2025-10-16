package main

import (
	"log"

	"github.com/gin-gonic/gin"
	ginbinding "github.com/zgs225/gin-form-binding"
)

func main() {
	// Create Gin router
	r := gin.Default()

	// Create form binding builder
	builder := ginbinding.NewBasicFormBindingGinHandlerBuilder(nil, nil)

	// Example 1: Simple handler with struct parameter
	userHandler := func(c *gin.Context, req struct {
		Name  string `json:"name" binding:"required"`
		Email string `json:"email" binding:"required,email"`
		Age   int    `json:"age" binding:"min=18"`
	}) (interface{}, error) {
		return gin.H{
			"message": "Hello " + req.Name,
			"email":   req.Email,
			"age":     req.Age,
		}, nil
	}

	userGinHandler, err := builder.FormBindingGinHandlerFunc(userHandler)
	if err != nil {
		log.Fatal(err)
	}

	// Example 2: Handler with path parameters
	profileHandler := func(c *gin.Context, req struct {
		UserID int    `path:"user_id"`
		Name   string `json:"name"`
		Bio    string `json:"bio"`
	}) (interface{}, error) {
		return gin.H{
			"user_id": req.UserID,
			"name":    req.Name,
			"bio":     req.Bio,
		}, nil
	}

	profileGinHandler, err := builder.FormBindingGinHandlerFunc(profileHandler)
	if err != nil {
		log.Fatal(err)
	}

	// Example 3: Handler with query parameters
	searchHandler := func(c *gin.Context, req struct {
		Query string `form:"q"`
		Page  int    `form:"page" default:"1"`
		Limit int    `form:"limit" default:"10"`
	}) (interface{}, error) {
		return gin.H{
			"query": req.Query,
			"page":  req.Page,
			"limit": req.Limit,
		}, nil
	}

	searchGinHandler, err := builder.FormBindingGinHandlerFunc(searchHandler)
	if err != nil {
		log.Fatal(err)
	}

	// Example 4: Handler without struct parameter
	healthHandler := func(c *gin.Context) (interface{}, error) {
		return gin.H{
			"status":  "healthy",
			"message": "Service is running",
		}, nil
	}

	healthGinHandler, err := builder.FormBindingGinHandlerFunc(healthHandler)
	if err != nil {
		log.Fatal(err)
	}

	// Register routes
	r.POST("/users", userGinHandler)
	r.PUT("/users/:user_id", profileGinHandler)
	r.GET("/search", searchGinHandler)
	r.GET("/health", healthGinHandler)

	// Start server
	log.Println("Server starting on :8080")
	log.Println("Try these endpoints:")
	log.Println("  POST /users - Create user")
	log.Println("  PUT /users/123 - Update user profile")
	log.Println("  GET /search?q=test&page=1&limit=5 - Search")
	log.Println("  GET /health - Health check")

	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
