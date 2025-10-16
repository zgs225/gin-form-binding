# Gin Form Binding

A flexible form binding middleware for the Gin framework that supports automatic parameter binding from path, query, header, and body with customizable response handling.

## Features

- **Multiple Function Signatures**: Support for three different handler function signatures
- **Comprehensive Binding**: Path parameters, query parameters, headers, and request body
- **Type Safety**: Full support for all Go primitive types, time.Time, time.Duration, and pointers
- **Default Values**: Support for default values via struct tags
- **Customizable Responses**: Pluggable response handling interface
- **Validation Integration**: Works with Gin's built-in validation
- **Zero Dependencies**: Only depends on Gin framework

## Installation

```bash
go get github.com/zgs225/gin-form-binding
```

## Quick Start

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/zgs225/gin-form-binding"
)

func main() {
    r := gin.Default()
    
    // Create a form binding builder
    builder := ginbinding.NewBasicFormBindingGinHandlerBuilder(nil, nil)
    
    // Define your handler function
    handler := func(c *gin.Context, req struct {
        Name  string `json:"name" binding:"required"`
        Email string `json:"email" binding:"required,email"`
        Age   int    `json:"age" binding:"min=18"`
    }) (interface{}, error) {
        // Your business logic here
        return gin.H{
            "message": "Hello " + req.Name,
            "email":   req.Email,
            "age":     req.Age,
        }, nil
    }
    
    // Convert to Gin handler
    ginHandler, err := builder.FormBindingGinHandlerFunc(handler)
    if err != nil {
        panic(err)
    }
    
    r.POST("/user", ginHandler)
    r.Run(":8080")
}
```

## Supported Function Signatures

The library supports three different function signatures:

### 1. Function with struct parameter returning error
```go
func(c *gin.Context, req struct {
    Name string `json:"name"`
}) error {
    // Process request
    return nil
}
```

### 2. Function with struct parameter returning data and error
```go
func(c *gin.Context, req struct {
    Name string `json:"name"`
}) (interface{}, error) {
    // Process request
    return gin.H{"message": "Hello " + req.Name}, nil
}
```

### 3. Function without struct parameter returning data and error
```go
func(c *gin.Context) (interface{}, error) {
    // Process request without binding
    return gin.H{"message": "Hello World"}, nil
}
```

## Supported Tags

### Path Parameters
```go
type Request struct {
    ID   int    `path:"id"`
    Name string `path:"name"`
}
```

### Query Parameters
```go
type Request struct {
    Page     int    `form:"page"`
    PageSize int    `form:"page_size"`
    Search   string `form:"search"`
}
```

### Headers
```go
type Request struct {
    UserAgent string `header:"User-Agent"`
    AuthToken string `header:"Authorization"`
}
```

### Request Body
```go
type Request struct {
    Name  string `json:"name"`
    Email string `json:"email"`
    Age   int    `json:"age"`
}
```

### Default Values
```go
type Request struct {
    Name     string        `json:"name" default:"John"`
    Age      int           `json:"age" default:"25"`
    IsActive bool          `json:"is_active" default:"true"`
    Timeout  time.Duration `json:"timeout" default:"30s"`
}
```

## Supported Data Types

- **Strings**: `string`
- **Integers**: `int`, `int8`, `int16`, `int32`, `int64`
- **Unsigned Integers**: `uint`, `uint8`, `uint16`, `uint32`, `uint64`
- **Floats**: `float32`, `float64`
- **Booleans**: `bool` (supports: true/false, 1/0, yes/no, on/off)
- **Time**: `time.Time` (multiple formats supported)
- **Duration**: `time.Duration`
- **Pointers**: All types can be pointers (`*string`, `*int`, etc.)

## Custom Response Handlers

You can provide custom response handling by implementing the `ResponseHandler` interface:

```go
type CustomResponseHandler struct{}

func (h *CustomResponseHandler) HandleSuccess(ctx *gin.Context, data interface{}) {
    ctx.JSON(http.StatusOK, gin.H{
        "success": true,
        "data":    data,
    })
}

func (h *CustomResponseHandler) HandleError(ctx *gin.Context, err error) {
    ctx.JSON(http.StatusBadRequest, gin.H{
        "success": false,
        "error":   err.Error(),
    })
}

// Use custom handler
builder := ginbinding.NewBasicFormBindingGinHandlerBuilder(nil, &CustomResponseHandler{})
```

## Validation Integration

The library works seamlessly with Gin's validation system:

```go
import "github.com/gin-gonic/gin/binding"

// Use Gin's default validator
builder := ginbinding.NewBasicFormBindingGinHandlerBuilder(binding.Validator, nil)

// Your handler with validation tags
handler := func(c *gin.Context, req struct {
    Name  string `json:"name" binding:"required"`
    Email string `json:"email" binding:"required,email"`
    Age   int    `json:"age" binding:"min=18,max=100"`
}) (interface{}, error) {
    return gin.H{"message": "Valid data"}, nil
}
```

## Advanced Examples

### Mixed Binding (Path + Query + Header + Body)
```go
handler := func(c *gin.Context, req struct {
    // Path parameter
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
```

### Embedded Structs
```go
type Pagination struct {
    Page     int `form:"page" default:"1"`
    PageSize int `form:"page_size" default:"10"`
}

type SearchRequest struct {
    Pagination
    Query string `form:"query"`
}

handler := func(c *gin.Context, req SearchRequest) (interface{}, error) {
    return gin.H{
        "page":      req.Page,
        "page_size": req.PageSize,
        "query":     req.Query,
    }, nil
}
```

### Default Values with Mixed Types
```go
handler := func(c *gin.Context, req struct {
    Name     string        `json:"name" default:"John"`
    Age      int           `json:"age" default:"25"`
    IsActive bool          `json:"is_active" default:"true"`
    Timeout  time.Duration `json:"timeout" default:"30s"`
    Created  time.Time     `json:"created" default:"2023-01-01T00:00:00Z"`
}) (interface{}, error) {
    return gin.H{
        "name":     req.Name,
        "age":      req.Age,
        "is_active": req.IsActive,
        "timeout":  req.Timeout.String(),
        "created":  req.Created.Format(time.RFC3339),
    }, nil
}
```

## Error Handling

The library provides comprehensive error handling:

- **Binding Errors**: Invalid type conversions, missing required fields
- **Validation Errors**: Custom validation failures
- **Handler Errors**: Errors returned from your handler functions

All errors are passed to your `ResponseHandler` for custom formatting.

## API Reference

### Types

```go
// ResponseHandler defines the interface for handling HTTP responses
type ResponseHandler interface {
    HandleSuccess(ctx *gin.Context, data interface{})
    HandleError(ctx *gin.Context, err error)
}

// FormBindingGinHandlerBuilder defines the interface for creating Gin handlers
type FormBindingGinHandlerBuilder interface {
    FormBindingGinHandlerFunc(i any) (gin.HandlerFunc, error)
}
```

### Functions

```go
// NewBasicFormBindingGinHandlerBuilder creates a new builder
func NewBasicFormBindingGinHandlerBuilder(
    validator binding.StructValidator,
    responseHandler ResponseHandler,
) *BasicFormBindingGinHandlerBuilder

// NewDefaultResponseHandler creates a default response handler
func NewDefaultResponseHandler() *DefaultResponseHandler
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Changelog

### v1.0.0
- Initial release
- Support for all three function signatures
- Comprehensive type support
- Default value handling
- Custom response handlers
- Validation integration
