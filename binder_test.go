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

func TestFormBindingGinHandlerFunc_ValidSignatures(t *testing.T) {
	builder := NewBasicFormBindingGinHandlerBuilder(nil, nil)

	tests := []struct {
		name     string
		handler  interface{}
		expected bool
	}{
		{
			name: "func(*gin.Context, struct) error",
			handler: func(c *gin.Context, req struct {
				Name string `json:"name"`
			}) error {
				return nil
			},
			expected: true,
		},
		{
			name: "func(*gin.Context, struct) (any, error)",
			handler: func(c *gin.Context, req struct {
				Name string `json:"name"`
			}) (interface{}, error) {
				return "success", nil
			},
			expected: true,
		},
		{
			name: "func(*gin.Context) (any, error)",
			handler: func(c *gin.Context) (interface{}, error) {
				return "success", nil
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, err := builder.FormBindingGinHandlerFunc(tt.handler)
			if tt.expected {
				assert.NoError(t, err)
				assert.NotNil(t, handler)
			} else {
				assert.Error(t, err)
				assert.Nil(t, handler)
			}
		})
	}
}

func TestFormBindingGinHandlerFunc_InvalidSignatures(t *testing.T) {
	builder := NewBasicFormBindingGinHandlerBuilder(nil, nil)

	tests := []struct {
		name     string
		handler  interface{}
		expected string
	}{
		{
			name:     "not a function",
			handler:  "not a function",
			expected: "input must be a function",
		},
		{
			name: "no parameters",
			handler: func() error {
				return nil
			},
			expected: "function must have at least one parameter",
		},
		{
			name: "too many parameters",
			handler: func(c *gin.Context, req struct{}, extra interface{}) error {
				return nil
			},
			expected: "function can have at most 2 parameters",
		},
		{
			name: "no return values",
			handler: func(c *gin.Context) {
			},
			expected: "function must have at least one return value",
		},
		{
			name: "too many return values",
			handler: func(c *gin.Context) (interface{}, interface{}, error) {
				return nil, nil, nil
			},
			expected: "function can have at most 2 return values",
		},
		{
			name: "first parameter not *gin.Context",
			handler: func(c string) error {
				return nil
			},
			expected: "first parameter must be *gin.Context",
		},
		{
			name: "second parameter not struct",
			handler: func(c *gin.Context, req string) error {
				return nil
			},
			expected: "second parameter must be a struct or pointer to struct",
		},
		{
			name: "single return not error",
			handler: func(c *gin.Context) string {
				return "test"
			},
			expected: "single return value must be error",
		},
		{
			name: "second return not error",
			handler: func(c *gin.Context) (interface{}, string) {
				return nil, "test"
			},
			expected: "second return value must be error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := builder.FormBindingGinHandlerFunc(tt.handler)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expected)
		})
	}
}

func TestHandlerExecution_WithStruct(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		handler        interface{}
		requestBody    string
		expectedStatus int
		expectedData   map[string]interface{}
	}{
		{
			name: "func(*gin.Context, struct) error - success",
			handler: func(c *gin.Context, req struct {
				Name string `json:"name"`
			}) error {
				if req.Name != "test" {
					return errors.New("invalid name")
				}
				return nil
			},
			requestBody:    `{"name":"test"}`,
			expectedStatus: http.StatusOK,
			expectedData:   map[string]interface{}{"status": "success"},
		},
		{
			name: "func(*gin.Context, struct) error - error",
			handler: func(c *gin.Context, req struct {
				Name string `json:"name"`
			}) error {
				return errors.New("validation failed")
			},
			requestBody:    `{"name":"test"}`,
			expectedStatus: http.StatusInternalServerError,
			expectedData:   map[string]interface{}{"status": "error", "message": "validation failed"},
		},
		{
			name: "func(*gin.Context, struct) (any, error) - success",
			handler: func(c *gin.Context, req struct {
				Name string `json:"name"`
			}) (interface{}, error) {
				return gin.H{"message": "Hello " + req.Name}, nil
			},
			requestBody:    `{"name":"test"}`,
			expectedStatus: http.StatusOK,
			expectedData:   map[string]interface{}{"status": "success", "data": map[string]interface{}{"message": "Hello test"}},
		},
		{
			name: "func(*gin.Context, struct) (any, error) - error",
			handler: func(c *gin.Context, req struct {
				Name string `json:"name"`
			}) (interface{}, error) {
				return nil, errors.New("processing failed")
			},
			requestBody:    `{"name":"test"}`,
			expectedStatus: http.StatusInternalServerError,
			expectedData:   map[string]interface{}{"status": "error", "message": "processing failed"},
		},
		{
			name: "func(*gin.Context) (any, error) - success",
			handler: func(c *gin.Context) (interface{}, error) {
				return gin.H{"message": "Hello World"}, nil
			},
			requestBody:    ``,
			expectedStatus: http.StatusOK,
			expectedData:   map[string]interface{}{"status": "success", "data": map[string]interface{}{"message": "Hello World"}},
		},
		{
			name: "func(*gin.Context) (any, error) - error",
			handler: func(c *gin.Context) (interface{}, error) {
				return nil, errors.New("service unavailable")
			},
			requestBody:    ``,
			expectedStatus: http.StatusInternalServerError,
			expectedData:   map[string]interface{}{"status": "error", "message": "service unavailable"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewBasicFormBindingGinHandlerBuilder(nil, nil)
			handler, err := builder.FormBindingGinHandlerFunc(tt.handler)
			assert.NoError(t, err)

			router := gin.New()
			router.POST("/test", handler)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/test", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			// Parse response body for comparison
			var response map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			// Compare the response structure
			assert.Equal(t, tt.expectedData["status"], response["status"])
			if tt.expectedData["message"] != nil {
				assert.Equal(t, tt.expectedData["message"], response["message"])
			}
			if tt.expectedData["data"] != nil {
				assert.NotNil(t, response["data"])
			}
		})
	}
}

func TestCustomResponseHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Custom response handler for testing
	customHandler := &testResponseHandler{}

	builder := NewBasicFormBindingGinHandlerBuilder(nil, customHandler)

	handler := func(c *gin.Context, req struct {
		Name string `json:"name"`
	}) (interface{}, error) {
		return gin.H{"message": "Hello " + req.Name}, nil
	}

	ginHandler, err := builder.FormBindingGinHandlerFunc(handler)
	assert.NoError(t, err)

	router := gin.New()
	router.POST("/test", ginHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test", strings.NewReader(`{"name":"test"}`))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "custom success", w.Body.String())
}

// testResponseHandler is a custom response handler for testing
type testResponseHandler struct{}

func (h *testResponseHandler) HandleSuccess(ctx *gin.Context, data interface{}) {
	ctx.String(http.StatusOK, "custom success")
}

func (h *testResponseHandler) HandleError(ctx *gin.Context, err error) {
	ctx.String(http.StatusBadRequest, "custom error: "+err.Error())
}
