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

func TestStringTypeConversion(t *testing.T) {
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

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test", strings.NewReader(`{"name":"John"}`))
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
}

func TestIntegerTypeConversion(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := func(c *gin.Context, req struct {
		IntVal   int     `json:"int_val"`
		FloatVal float64 `json:"float_val"`
	}) (interface{}, error) {
		return gin.H{
			"int_val":   req.IntVal,
			"float_val": req.FloatVal,
		}, nil
	}

	builder := NewBasicFormBindingGinHandlerBuilder(nil, nil)
	ginHandler, err := builder.FormBindingGinHandlerFunc(handler)
	assert.NoError(t, err)

	router := gin.New()
	router.POST("/test", ginHandler)

	// Test with valid integers
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test", strings.NewReader(`{"int_val":42,"float_val":3.14}`))
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
	// All numbers will be float64 due to JSON unmarshaling
	assert.Equal(t, float64(42), data["int_val"])
	assert.Equal(t, 3.14, data["float_val"])
}

func TestBooleanTypeConversion(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := func(c *gin.Context, req struct {
		BoolVal bool `json:"bool_val"`
	}) (interface{}, error) {
		return gin.H{"bool_val": req.BoolVal}, nil
	}

	builder := NewBasicFormBindingGinHandlerBuilder(nil, nil)
	ginHandler, err := builder.FormBindingGinHandlerFunc(handler)
	assert.NoError(t, err)

	router := gin.New()
	router.POST("/test", ginHandler)

	tests := []struct {
		name     string
		body     string
		expected bool
	}{
		{"true value", `{"bool_val":true}`, true},
		{"false value", `{"bool_val":false}`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/test", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			// Check basic structure
			assert.Equal(t, "success", response["status"])
			assert.NotNil(t, response["data"])

			data := response["data"].(map[string]interface{})
			assert.Equal(t, tt.expected, data["bool_val"])
		})
	}
}

func TestZeroValues(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := func(c *gin.Context, req struct {
		StringVal string `json:"string_val"`
		IntVal    int    `json:"int_val"`
		BoolVal   bool   `json:"bool_val"`
	}) (interface{}, error) {
		return gin.H{
			"string_val": req.StringVal,
			"int_val":    req.IntVal,
			"bool_val":   req.BoolVal,
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

	// Check basic structure
	assert.Equal(t, "success", response["status"])
	assert.NotNil(t, response["data"])

	data := response["data"].(map[string]interface{})
	assert.Equal(t, "", data["string_val"])
	assert.Equal(t, float64(0), data["int_val"]) // JSON numbers are float64
	assert.Equal(t, false, data["bool_val"])
}
