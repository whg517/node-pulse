package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Error codes
const (
	ERR_INVALID_REQUEST = "ERR_INVALID_REQUEST"
	ERR_DATABASE_ERROR = "ERR_DATABASE_ERROR"
	ERR_INTERNAL       = "ERR_INTERNAL"
	ERR_NOT_FOUND      = "ERR_NOT_FOUND"
	ERR_UNAUTHORIZED   = "ERR_UNAUTHORIZED"
)

// ErrorResponse represents standard error response format
type ErrorResponse struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// Error implements the error interface
func (e *ErrorResponse) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// AppError represents an application error
type AppError struct {
	Code    string
	Message string
}

// Error implements the error interface
func (e *AppError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// ErrorHandler middleware handles errors and returns standardized error responses
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if there were any errors
		if len(c.Errors) > 0 {
			lastError := c.Errors.Last()
			err := lastError.Err

			// Determine HTTP status code and error code based on error type
			var httpStatus int
			var errorCode string
			var message string

			switch v := err.(type) {
			case *ErrorResponse:
				errorCode = v.Code
				message = v.Message
				httpStatus = http.StatusInternalServerError
			case *AppError:
				errorCode = v.Code
				message = v.Message
				httpStatus = http.StatusInternalServerError
			case error:
				errorCode = ERR_INTERNAL
				message = v.Error()
				httpStatus = http.StatusInternalServerError
			default:
				errorCode = ERR_INTERNAL
				message = "An unexpected error occurred"
				httpStatus = http.StatusInternalServerError
			}

			c.JSON(httpStatus, ErrorResponse{
				Code:    errorCode,
				Message: message,
			})
		}
	}
}

// RespondWithError sends a standardized error response
// This should only be used when you want to immediately respond with an error
// and NOT use it when you want the ErrorHandler middleware to catch errors
func RespondWithError(c *gin.Context, code string, message string, statusCode int) {
	c.JSON(statusCode, ErrorResponse{
		Code:    code,
		Message: message,
	})
	// Abort to prevent further handlers
	c.Abort()
}

// RespondWithSuccess sends a standardized success response
func RespondWithSuccess(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"code":    "SUCCESS",
		"message": "OK",
		"data":    data,
	})
}
