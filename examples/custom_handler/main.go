package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	ginbinding "github.com/zgs225/gin-form-binding"
)

// Custom response handler that returns XML instead of JSON
type XMLResponseHandler struct{}

func (h *XMLResponseHandler) HandleSuccess(ctx *gin.Context, data interface{}) {
	ctx.Header("Content-Type", "application/xml")
	ctx.String(http.StatusOK, `<?xml version="1.0" encoding="UTF-8"?>
<response>
    <success>true</success>
    <data>%v</data>
</response>`, data)
}

func (h *XMLResponseHandler) HandleError(ctx *gin.Context, err error) {
	ctx.Header("Content-Type", "application/xml")
	ctx.String(http.StatusBadRequest, `<?xml version="1.0" encoding="UTF-8"?>
<response>
    <success>false</success>
    <error>%s</error>
</response>`, err.Error())
}

// Custom response handler that returns plain text
type TextResponseHandler struct{}

func (h *TextResponseHandler) HandleSuccess(ctx *gin.Context, data interface{}) {
	ctx.String(http.StatusOK, "Success: %v", data)
}

func (h *TextResponseHandler) HandleError(ctx *gin.Context, err error) {
	ctx.String(http.StatusBadRequest, "Error: %s", err.Error())
}

func main() {
	// Create Gin router
	r := gin.Default()

	// Example 1: XML response handler
	xmlBuilder := ginbinding.NewBasicFormBindingGinHandlerBuilder(nil, &XMLResponseHandler{})

	xmlHandler := func(c *gin.Context, req struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}) (interface{}, error) {
		return gin.H{
			"name":  req.Name,
			"email": req.Email,
		}, nil
	}

	xmlGinHandler, err := xmlBuilder.FormBindingGinHandlerFunc(xmlHandler)
	if err != nil {
		log.Fatal(err)
	}

	// Example 2: Text response handler
	textBuilder := ginbinding.NewBasicFormBindingGinHandlerBuilder(nil, &TextResponseHandler{})

	textHandler := func(c *gin.Context, req struct {
		Message string `json:"message"`
	}) (interface{}, error) {
		return "Processed: " + req.Message, nil
	}

	textGinHandler, err := textBuilder.FormBindingGinHandlerFunc(textHandler)
	if err != nil {
		log.Fatal(err)
	}

	// Register routes
	r.POST("/xml", xmlGinHandler)
	r.POST("/text", textGinHandler)

	// Start server
	log.Println("Custom handler server starting on :8080")
	log.Println("Try these endpoints:")
	log.Println("  POST /xml - Returns XML response")
	log.Println("  POST /text - Returns plain text response")

	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
