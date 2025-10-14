package utils

import (
	"log"

	"github.com/gin-gonic/gin"
)

// ErrorResponse represents an error response
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}

// SuccessResponse represents a success response
type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// SendErrorResponse sends an error response
func SendErrorResponse(c *gin.Context, statusCode int, message string, err error) {
	response := ErrorResponse{
		Success: false,
		Error:   message,
	}

	if err != nil {
		response.Details = err.Error()
		log.Printf("Error: %s - %v", message, err)
	}

	c.JSON(statusCode, response)
}

// SendSuccessResponse sends a success response
func SendSuccessResponse(c *gin.Context, statusCode int, data interface{}, message string) {
	response := SuccessResponse{
		Success: true,
		Data:    data,
		Message: message,
	}

	c.JSON(statusCode, response)
}

// StringPtr returns a pointer to the given string
func StringPtr(s string) *string {
	return &s
}
