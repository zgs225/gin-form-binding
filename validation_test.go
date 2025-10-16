package ginbinding

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Mock validator for testing
type mockValidator struct {
	shouldError bool
	errorMsg    string
}

func (m *mockValidator) ValidateStruct(obj interface{}) error {
	if m.shouldError {
		return errors.New("validation failed")
	}
	return nil
}

func (m *mockValidator) Engine() interface{} {
	return nil
}

func TestValidationIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := func(c *gin.Context, req struct {
		Name  string `json:"name" binding:"required"`
		Email string `json:"email" binding:"required,email"`
		Age   int    `json:"age" binding:"min=18,max=100"`
	}) (interface{}, error) {
		return gin.H{
			"name":  req.Name,
			"email": req.Email,
			"age":   req.Age,
		}, nil
	}

	// Test with validator that passes
	validator := &mockValidator{shouldError: false}
	builder := NewBasicFormBindingGinHandlerBuilder(validator, nil)
	ginHandler, err := builder.FormBindingGinHandlerFunc(handler)
	assert.NoError(t, err)

	router := gin.New()
	router.POST("/test", ginHandler)

	// Test valid data
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test", strings.NewReader(`{"name":"John","email":"john@example.com","age":25}`))
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
	assert.Equal(t, float64(25), data["age"]) // JSON numbers are float64
}

func TestValidationFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := func(c *gin.Context, req struct {
		Name  string `json:"name" binding:"required"`
		Email string `json:"email" binding:"required,email"`
		Age   int    `json:"age" binding:"min=18,max=100"`
	}) (interface{}, error) {
		return gin.H{
			"name":  req.Name,
			"email": req.Email,
			"age":   req.Age,
		}, nil
	}

	// Test with validator that fails
	validator := &mockValidator{shouldError: true}
	builder := NewBasicFormBindingGinHandlerBuilder(validator, nil)
	ginHandler, err := builder.FormBindingGinHandlerFunc(handler)
	assert.NoError(t, err)

	router := gin.New()
	router.POST("/test", ginHandler)

	// Test with valid data but validator fails
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test", strings.NewReader(`{"name":"John","email":"john@example.com","age":25}`))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "error", response["status"])
	assert.Contains(t, response["message"], "validation")
}

func TestValidationWithoutValidator(t *testing.T) {
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

	// Test without validator (nil)
	builder := NewBasicFormBindingGinHandlerBuilder(nil, nil)
	ginHandler, err := builder.FormBindingGinHandlerFunc(handler)
	assert.NoError(t, err)

	router := gin.New()
	router.POST("/test", ginHandler)

	// Test with valid data - should work without validation
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test", strings.NewReader(`{"name":"John","email":"john@example.com","age":25}`))
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
	assert.Equal(t, float64(25), data["age"]) // JSON numbers are float64
}

func TestValidationWithCustomResponseHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := func(c *gin.Context, req struct {
		Name string `json:"name" binding:"required"`
	}) (interface{}, error) {
		return gin.H{"name": req.Name}, nil
	}

	// Custom response handler that captures validation errors
	customHandler := &testValidationResponseHandler{}

	validator := &mockValidator{shouldError: true}
	builder := NewBasicFormBindingGinHandlerBuilder(validator, customHandler)
	ginHandler, err := builder.FormBindingGinHandlerFunc(handler)
	assert.NoError(t, err)

	router := gin.New()
	router.POST("/test", ginHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test", strings.NewReader(`{"name":"John"}`))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "validation error", w.Body.String())
}

// testValidationResponseHandler is a custom response handler for testing validation
type testValidationResponseHandler struct{}

func (h *testValidationResponseHandler) HandleSuccess(ctx *gin.Context, data interface{}) {
	ctx.String(http.StatusOK, "success")
}

func (h *testValidationResponseHandler) HandleError(ctx *gin.Context, err error) {
	ctx.String(http.StatusBadRequest, "validation error")
}
