package ginbinding

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// DefaultResponseHandler provides a standard JSON response handler
type DefaultResponseHandler struct{}

// NewDefaultResponseHandler creates a new default response handler
func NewDefaultResponseHandler() *DefaultResponseHandler {
	return &DefaultResponseHandler{}
}

// HandleSuccess sends a JSON response with the provided data
func (h *DefaultResponseHandler) HandleSuccess(ctx *gin.Context, data interface{}) {
	if data == nil {
		ctx.JSON(http.StatusOK, gin.H{"status": "success"})
	} else {
		ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": data})
	}
}

// HandleError sends a JSON error response with appropriate HTTP status code
func (h *DefaultResponseHandler) HandleError(ctx *gin.Context, err error) {
	statusCode := http.StatusInternalServerError
	message := "Internal server error"

	// Check if it's a binding error
	if bindingErr, ok := err.(*BindingError); ok {
		statusCode = http.StatusBadRequest
		message = bindingErr.Error()
	} else {
		// For other errors, try to determine appropriate status code
		switch err.Error() {
		case "record not found":
			statusCode = http.StatusNotFound
			message = err.Error()
		case "unauthorized":
			statusCode = http.StatusUnauthorized
			message = err.Error()
		case "forbidden":
			statusCode = http.StatusForbidden
			message = err.Error()
		default:
			message = err.Error()
		}
	}

	ctx.JSON(statusCode, gin.H{
		"status":  "error",
		"message": message,
	})
}
