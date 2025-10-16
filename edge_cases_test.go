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

func TestEmptyRequestBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := func(c *gin.Context, req struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}) (interface{}, error) {
		return gin.H{
			"name": req.Name,
			"age":  req.Age,
		}, nil
	}

	builder := NewBasicFormBindingGinHandlerBuilder(nil, nil)
	ginHandler, err := builder.FormBindingGinHandlerFunc(handler)
	assert.NoError(t, err)

	router := gin.New()
	router.POST("/test", ginHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")

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
	assert.Equal(t, "", data["name"])
	assert.Equal(t, float64(0), data["age"]) // JSON numbers are float64
}

func TestNilRequestBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := func(c *gin.Context, req struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}) (interface{}, error) {
		return gin.H{
			"name": req.Name,
			"age":  req.Age,
		}, nil
	}

	builder := NewBasicFormBindingGinHandlerBuilder(nil, nil)
	ginHandler, err := builder.FormBindingGinHandlerFunc(handler)
	assert.NoError(t, err)

	router := gin.New()
	router.POST("/test", ginHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")

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
	assert.Equal(t, "", data["name"])
	assert.Equal(t, float64(0), data["age"]) // JSON numbers are float64
}

func TestInvalidTypeConversion(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := func(c *gin.Context, req struct {
		Age int `json:"age"`
	}) (interface{}, error) {
		return gin.H{"age": req.Age}, nil
	}

	builder := NewBasicFormBindingGinHandlerBuilder(nil, nil)
	ginHandler, err := builder.FormBindingGinHandlerFunc(handler)
	assert.NoError(t, err)

	router := gin.New()
	router.POST("/test", ginHandler)

	// Test with invalid integer
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test", strings.NewReader(`{"age":"not_a_number"}`))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "error", response["status"])
	// Check that we get an error message (the exact message may vary)
	assert.NotNil(t, response["message"])
	assert.NotEmpty(t, response["message"])
}

func TestUnexportedFields(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := func(c *gin.Context, req struct {
		Name    string `json:"name"`
		age     int    `json:"age"`     // unexported field
		private string `json:"private"` // unexported field
	}) (interface{}, error) {
		return gin.H{
			"name":    req.Name,
			"age":     req.age,
			"private": req.private,
		}, nil
	}

	builder := NewBasicFormBindingGinHandlerBuilder(nil, nil)
	ginHandler, err := builder.FormBindingGinHandlerFunc(handler)
	assert.NoError(t, err)

	router := gin.New()
	router.POST("/test", ginHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test", strings.NewReader(`{"name":"John","age":25,"private":"secret"}`))
	req.Header.Set("Content-Type", "application/json")

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
	assert.Equal(t, "John", data["name"])
	assert.Equal(t, float64(0), data["age"]) // unexported field should be zero value
	assert.Equal(t, "", data["private"])     // unexported field should be zero value
}

func TestEmbeddedAnonymousStructs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	type Base struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	handler := func(c *gin.Context, req struct {
		Base
		Email string `json:"email"`
	}) (interface{}, error) {
		return gin.H{
			"id":    req.ID,
			"name":  req.Name,
			"email": req.Email,
		}, nil
	}

	builder := NewBasicFormBindingGinHandlerBuilder(nil, nil)
	ginHandler, err := builder.FormBindingGinHandlerFunc(handler)
	assert.NoError(t, err)

	router := gin.New()
	router.POST("/test", ginHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test", strings.NewReader(`{"id":123,"name":"John","email":"john@example.com"}`))
	req.Header.Set("Content-Type", "application/json")

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
	assert.Equal(t, float64(123), data["id"]) // JSON numbers are float64
	assert.Equal(t, "John", data["name"])
	assert.Equal(t, "john@example.com", data["email"])
}

func TestPointerToStruct(t *testing.T) {
	gin.SetMode(gin.TestMode)

	type User struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	handler := func(c *gin.Context, req struct {
		User *User `json:"user"`
	}) (interface{}, error) {
		if req.User == nil {
			return gin.H{"user": nil}, nil
		}
		return gin.H{
			"user": gin.H{
				"name":  req.User.Name,
				"email": req.User.Email,
			},
		}, nil
	}

	builder := NewBasicFormBindingGinHandlerBuilder(nil, nil)
	ginHandler, err := builder.FormBindingGinHandlerFunc(handler)
	assert.NoError(t, err)

	router := gin.New()
	router.POST("/test", ginHandler)

	// Test with valid user data
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test", strings.NewReader(`{"user":{"name":"John","email":"john@example.com"}}`))
	req.Header.Set("Content-Type", "application/json")

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

	// Check user data
	assert.NotNil(t, data["user"])
	user := data["user"].(map[string]interface{})
	assert.Equal(t, "John", user["name"])
	assert.Equal(t, "john@example.com", user["email"])
}

func TestEmptyStringValues(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := func(c *gin.Context, req struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Age   int    `json:"age"`
	}) (interface{}, error) {
		return gin.H{
			"name":  req.Name,
			"email": req.Email,
			"age":   req.Age,
		}, nil
	}

	builder := NewBasicFormBindingGinHandlerBuilder(nil, nil)
	ginHandler, err := builder.FormBindingGinHandlerFunc(handler)
	assert.NoError(t, err)

	router := gin.New()
	router.POST("/test", ginHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test", strings.NewReader(`{"name":"","email":"","age":0}`))
	req.Header.Set("Content-Type", "application/json")

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
	assert.Equal(t, "", data["name"])
	assert.Equal(t, "", data["email"])
	assert.Equal(t, float64(0), data["age"]) // JSON numbers are float64
}

func TestInvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := func(c *gin.Context, req struct {
		Name string `json:"name"`
	}) (interface{}, error) {
		return gin.H{"name": req.Name}, nil
	}

	builder := NewBasicFormBindingGinHandlerBuilder(nil, nil)
	ginHandler, err := builder.FormBindingGinHandlerFunc(handler)
	assert.NoError(t, err)

	router := gin.New()
	router.POST("/test", ginHandler)

	// Test with invalid JSON
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test", strings.NewReader(`{"name":"John",}`)) // Invalid JSON
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "error", response["status"])
	// Check that we get an error message (the exact message may vary)
	assert.NotNil(t, response["message"])
	assert.NotEmpty(t, response["message"])
}

func TestHandlerFunctionError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := func(c *gin.Context, req struct {
		Name string `json:"name"`
	}) (interface{}, error) {
		return nil, assert.AnError
	}

	builder := NewBasicFormBindingGinHandlerBuilder(nil, nil)
	ginHandler, err := builder.FormBindingGinHandlerFunc(handler)
	assert.NoError(t, err)

	router := gin.New()
	router.POST("/test", ginHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test", strings.NewReader(`{"name":"John"}`))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "error", response["status"])
	assert.Contains(t, response["message"], "assert.AnError general error for testing")
}
