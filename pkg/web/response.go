package web

import "github.com/gin-gonic/gin"

// Response represents a standard response format
type Response struct {
	Status  string      `json:"status"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// JSON sends a successful response in JSON format
func JSON(c *gin.Context, statusCode int, data interface{}, message string) {
	c.JSON(statusCode, Response{
		Status:  "success",
		Data:    data,
		Message: message,
	})
}

// Error sends an error response in JSON format
func Error(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, Response{
		Status:  "error",
		Message: message,
	})
}

// PaginatedResponse represents a paginated data response
type PaginatedResponse struct {
	Items      interface{} `json:"items"`
	TotalCount int64       `json:"total_count"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
}

func PaginatedJSON(c *gin.Context, statusCode int, paginated PaginatedResponse, message string) {
	c.JSON(statusCode, Response{
		Status:  "success",
		Data:    paginated,
		Message: message,
	})
}
