package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"kirimaja-go/internal/common/response"
)

func RequirePermission(db *gorm.DB) func(permissionKey string) gin.HandlerFunc {
	return func(permissionKey string) gin.HandlerFunc {
		return func(c *gin.Context) {
			userID, exists := c.Get("userID")
			if !exists {
				response.Error(c, http.StatusUnauthorized, "Unauthorized", nil)
				c.Abort()
				return
			}

			var count int64
			err := db.Raw(`
				SELECT COUNT(*) FROM permissions p
				JOIN role_permissions rp ON rp.permission_id = p.id
				JOIN roles r ON r.id = rp.role_id
				JOIN users u ON u.role_id = r.id
				WHERE u.id = ? AND p.key = ?
			`, userID, permissionKey).Scan(&count).Error

			if err != nil {
				// A DB outage is not an authorization decision — surface it
				// as 500 so it isn't silently mistaken for "access denied".
				response.Error(c, http.StatusInternalServerError, "Failed to verify permissions", nil)
				c.Abort()
				return
			}
			if count == 0 {
				response.Error(c, http.StatusForbidden, "Access denied: insufficient permissions", nil)
				c.Abort()
				return
			}
			c.Next()
		}
	}
}
