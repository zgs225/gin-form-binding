package ginbinding

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestPathParameterBinding(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := func(c *gin.Context, req struct {
		ID   int    `path:"id"`
		Name string `path:"name"`
	}) (interface{}, error) {
		return gin.H{"id": req.ID, "name": req.Name}, nil
	}

	builder := NewBasicFormBindingGinHandlerBuilder(nil, nil)
	ginHandler, err := builder.FormBindingGinHandlerFunc(handler)
	assert.NoError(t, err)

	router := gin.New()
	router.GET("/user/:id/:name", ginHandler)

	// Test valid path parameters
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/user/123/john", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Check basic structure
	assert.Equal(t, "success", response["status"])
	assert.NotNil(t, response["data"])

	data := response["data"].(map[string]interface{})
	assert.Equal(t, "john", data["name"])
	// ID will be float64 due to JSON unmarshaling
	assert.Equal(t, float64(123), data["id"])
}

func TestQueryParameterBinding(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := func(c *gin.Context, req struct {
		Page     int    `form:"page"`
		PageSize int    `form:"page_size"`
		Search   string `form:"search"`
		Active   bool   `form:"active"`
	}) (interface{}, error) {
		return gin.H{
			"page":      req.Page,
			"page_size": req.PageSize,
			"search":    req.Search,
			"active":    req.Active,
		}, nil
	}

	builder := NewBasicFormBindingGinHandlerBuilder(nil, nil)
	ginHandler, err := builder.FormBindingGinHandlerFunc(handler)
	assert.NoError(t, err)

	router := gin.New()
	router.GET("/search", ginHandler)

	// Test valid query parameters
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/search?page=1&page_size=10&search=test&active=true", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Check basic structure
	assert.Equal(t, "success", response["status"])
	assert.NotNil(t, response["data"])

	data := response["data"].(map[string]interface{})
	assert.Equal(t, "test", data["search"])
	assert.Equal(t, true, data["active"])
	// Numbers will be float64 due to JSON unmarshaling
	assert.Equal(t, float64(1), data["page"])
	assert.Equal(t, float64(10), data["page_size"])
}

func TestHeaderBinding(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := func(c *gin.Context, req struct {
		UserAgent   string `header:"User-Agent"`
		AuthToken   string `header:"Authorization"`
		ContentType string `header:"Content-Type"`
	}) (interface{}, error) {
		return gin.H{
			"user_agent":   req.UserAgent,
			"auth_token":   req.AuthToken,
			"content_type": req.ContentType,
		}, nil
	}

	builder := NewBasicFormBindingGinHandlerBuilder(nil, nil)
	ginHandler, err := builder.FormBindingGinHandlerFunc(handler)
	assert.NoError(t, err)

	router := gin.New()
	router.POST("/headers", ginHandler)

	// Test with headers
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/headers", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "test-agent")
	req.Header.Set("Authorization", "Bearer token123")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Check basic structure
	assert.Equal(t, "success", response["status"])
	assert.NotNil(t, response["data"])

	data := response["data"].(map[string]interface{})
	assert.Equal(t, "test-agent", data["user_agent"])
	assert.Equal(t, "Bearer token123", data["auth_token"])
	assert.Equal(t, "application/json", data["content_type"])
}

func TestRequestBodyBinding(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := func(c *gin.Context, req struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Age      int    `json:"age"`
		IsActive bool   `json:"is_active"`
	}) (interface{}, error) {
		return gin.H{
			"name":      req.Name,
			"email":     req.Email,
			"age":       req.Age,
			"is_active": req.IsActive,
		}, nil
	}

	builder := NewBasicFormBindingGinHandlerBuilder(nil, nil)
	ginHandler, err := builder.FormBindingGinHandlerFunc(handler)
	assert.NoError(t, err)

	router := gin.New()
	router.POST("/user", ginHandler)

	// Test valid JSON body
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/user", strings.NewReader(`{"name":"John","email":"john@example.com","age":30,"is_active":true}`))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Check basic structure
	assert.Equal(t, "success", response["status"])
	assert.NotNil(t, response["data"])

	data := response["data"].(map[string]interface{})
	assert.Equal(t, "John", data["name"])
	assert.Equal(t, "john@example.com", data["email"])
	assert.Equal(t, true, data["is_active"])
	// Age will be float64 due to JSON unmarshaling
	assert.Equal(t, float64(30), data["age"])
}

func TestMixedBinding(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := func(c *gin.Context, req struct {
		// Path parameters
		UserID int `path:"user_id"`

		// Query parameters
		Page int `form:"page"`

		// Headers
		AuthToken string `header:"Authorization"`

		// Body
		Name  string `json:"name"`
		Email string `json:"email"`
	}) (interface{}, error) {
		return gin.H{
			"user_id":    req.UserID,
			"page":       req.Page,
			"auth_token": req.AuthToken,
			"name":       req.Name,
			"email":      req.Email,
		}, nil
	}

	builder := NewBasicFormBindingGinHandlerBuilder(nil, nil)
	ginHandler, err := builder.FormBindingGinHandlerFunc(handler)
	assert.NoError(t, err)

	router := gin.New()
	router.PUT("/user/:user_id", ginHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/user/123?page=1", strings.NewReader(`{"name":"John","email":"john@example.com"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token123")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Check status
	assert.Equal(t, "success", response["status"])

	// Check data
	assert.NotNil(t, response["data"])
	data := response["data"].(map[string]interface{})

	// Compare specific fields
	assert.Equal(t, float64(123), data["user_id"]) // JSON numbers are float64
	assert.Equal(t, float64(1), data["page"])      // JSON numbers are float64
	assert.Equal(t, "Bearer token123", data["auth_token"])
	assert.Equal(t, "John", data["name"])
	assert.Equal(t, "john@example.com", data["email"])
}
