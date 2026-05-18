package response

import "github.com/gin-gonic/gin"

type Meta struct {
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
	Total   int `json:"total"`
}

func Success(c *gin.Context, status int, message string, data any) {
	c.JSON(status, gin.H{
		"success": true,
		"message": message,
		"data":    data,
	})
}

func Error(c *gin.Context, status int, message string, err any) {
	c.JSON(status, gin.H{
		"success": false,
		"message": message,
		"error":   err,
	})
}

func Paginated(c *gin.Context, status int, message string, data any, meta Meta) {
	c.JSON(status, gin.H{
		"success": true,
		"message": message,
		"data":    data,
		"meta":    meta,
	})
}
