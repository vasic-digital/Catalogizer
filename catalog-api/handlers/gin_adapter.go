package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// WrapHTTPHandler adapts a standard net/http handler to a gin.HandlerFunc.
// This allows handlers written for the standard library to be used with Gin routes.
func WrapHTTPHandler(handler func(http.ResponseWriter, *http.Request)) gin.HandlerFunc {
	return func(c *gin.Context) {
		handler(c.Writer, c.Request)
	}
}
