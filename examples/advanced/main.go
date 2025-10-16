package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	ginbinding "github.com/zgs225/gin-form-binding"
)

// Custom response handler for advanced use cases
type CustomResponseHandler struct{}

func (h *CustomResponseHandler) HandleSuccess(ctx *gin.Context, data interface{}) {
	ctx.JSON(http.StatusOK, gin.H{
		"success":   true,
		"data":      data,
		"timestamp": time.Now().Unix(),
	})
}

func (h *CustomResponseHandler) HandleError(ctx *gin.Context, err error) {
	ctx.JSON(http.StatusBadRequest, gin.H{
		"success":   false,
		"error":     err.Error(),
		"timestamp": time.Now().Unix(),
	})
}

// Embedded struct for pagination
type Pagination struct {
	Page     int `form:"page" default:"1"`
	PageSize int `form:"page_size" default:"10"`
}

// Advanced request struct with mixed binding sources
type AdvancedRequest struct {
	Pagination
	UserID    int           `path:"user_id"`
	AuthToken string        `header:"Authorization"`
	Name      string        `json:"name" binding:"required"`
	Email     string        `json:"email" binding:"required,email"`
	Age       int           `json:"age" binding:"min=18,max=100"`
	IsActive  bool          `json:"is_active" default:"true"`
	Timeout   time.Duration `json:"timeout" default:"30s"`
	Created   time.Time     `json:"created" default:"2023-01-01T00:00:00Z"`
}

func main() {
	// Create Gin router
	r := gin.Default()

	// Create form binding builder with custom response handler and validation
	builder := ginbinding.NewBasicFormBindingGinHandlerBuilder(binding.Validator, &CustomResponseHandler{})

	// Advanced handler with mixed binding
	advancedHandler := func(c *gin.Context, req AdvancedRequest) (interface{}, error) {
		return gin.H{
			"user_id":    req.UserID,
			"auth_token": req.AuthToken,
			"name":       req.Name,
			"email":      req.Email,
			"age":        req.Age,
			"is_active":  req.IsActive,
			"timeout":    req.Timeout.String(),
			"created":    req.Created.Format(time.RFC3339),
			"pagination": gin.H{
				"page":      req.Page,
				"page_size": req.PageSize,
			},
		}, nil
	}

	advancedGinHandler, err := builder.FormBindingGinHandlerFunc(advancedHandler)
	if err != nil {
		log.Fatal(err)
	}

	// Handler with pointer types
	pointerHandler := func(c *gin.Context, req struct {
		Name  *string `json:"name"`
		Age   *int    `json:"age"`
		Email *string `json:"email"`
	}) (interface{}, error) {
		result := gin.H{}
		if req.Name != nil {
			result["name"] = *req.Name
		}
		if req.Age != nil {
			result["age"] = *req.Age
		}
		if req.Email != nil {
			result["email"] = *req.Email
		}
		return result, nil
	}

	pointerGinHandler, err := builder.FormBindingGinHandlerFunc(pointerHandler)
	if err != nil {
		log.Fatal(err)
	}

	// Handler with time parsing
	timeHandler := func(c *gin.Context, req struct {
		StartTime time.Time     `json:"start_time"`
		Duration  time.Duration `json:"duration"`
		EndTime   time.Time     `json:"end_time" default:"2023-12-31T23:59:59Z"`
	}) (interface{}, error) {
		return gin.H{
			"start_time": req.StartTime.Format(time.RFC3339),
			"duration":   req.Duration.String(),
			"end_time":   req.EndTime.Format(time.RFC3339),
		}, nil
	}

	timeGinHandler, err := builder.FormBindingGinHandlerFunc(timeHandler)
	if err != nil {
		log.Fatal(err)
	}

	// Handler with complex validation
	validationHandler := func(c *gin.Context, req struct {
		Username string `json:"username" binding:"required,min=3,max=20"`
		Password string `json:"password" binding:"required,min=8"`
		Email    string `json:"email" binding:"required,email"`
		Age      int    `json:"age" binding:"required,min=18,max=100"`
	}) (interface{}, error) {
		return gin.H{
			"username": req.Username,
			"email":    req.Email,
			"age":      req.Age,
			"message":  "Validation passed",
		}, nil
	}

	validationGinHandler, err := builder.FormBindingGinHandlerFunc(validationHandler)
	if err != nil {
		log.Fatal(err)
	}

	// Register routes
	r.PUT("/users/:user_id", advancedGinHandler)
	r.POST("/pointer-test", pointerGinHandler)
	r.POST("/time-test", timeGinHandler)
	r.POST("/validation-test", validationGinHandler)

	// Start server
	log.Println("Advanced server starting on :8080")
	log.Println("Try these endpoints:")
	log.Println("  PUT /users/123 - Advanced mixed binding")
	log.Println("  POST /pointer-test - Pointer types")
	log.Println("  POST /time-test - Time parsing")
	log.Println("  POST /validation-test - Complex validation")

	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
