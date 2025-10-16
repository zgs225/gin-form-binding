// Package ginbinding provides a flexible form binding middleware for Gin framework
// that supports automatic parameter binding from path, query, header, and body
// with customizable response handling.
package ginbinding

import (
	"github.com/gin-gonic/gin"
)

// ResponseHandler defines the interface for handling HTTP responses
// in the form binding middleware.
type ResponseHandler interface {
	// HandleSuccess is called when the handler function executes successfully
	// data can be nil for functions that return only error
	HandleSuccess(ctx *gin.Context, data interface{})

	// HandleError is called when an error occurs during binding or handler execution
	HandleError(ctx *gin.Context, err error)
}

// FormBindingGinHandlerBuilder defines the interface for creating Gin handlers
// from functions that match the supported signatures.
type FormBindingGinHandlerBuilder interface {
	// FormBindingGinHandlerFunc converts a function to a gin.HandlerFunc
	// Supported function signatures:
	//  1. func(*gin.Context, any struct) error
	//  2. func(*gin.Context, any struct) (any, error)
	//  3. func(*gin.Context) (any, error)
	FormBindingGinHandlerFunc(i any) (gin.HandlerFunc, error)
}

// BindingError represents an error that occurred during form binding
type BindingError struct {
	Err error
}

// Error implements the error interface
func (e *BindingError) Error() string {
	return e.Err.Error()
}

// Unwrap returns the underlying error
func (e *BindingError) Unwrap() error {
	return e.Err
}
