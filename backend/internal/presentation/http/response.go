package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Success writes a JSON success response.
func Success(c *gin.Context, status int, data interface{}) {
	c.JSON(status, data)
}

// Error writes a JSON error response.
func Error(c *gin.Context, status int, message string) {
	c.AbortWithStatusJSON(status, gin.H{
		"error":   http.StatusText(status),
		"message": message,
	})
}

// Paginated writes a paginated list response.
func Paginated(c *gin.Context, data interface{}, total, page, pageSize, totalPages int) {
	c.JSON(http.StatusOK, gin.H{
		"data":        data,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	})
}
