package permissions

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.RouterGroup, h *Handler, authMw gin.HandlerFunc, permMw func(string) gin.HandlerFunc) {
	g := r.Group("/permissions", authMw)
	{
		g.GET("", permMw("permission:read"), h.FindAll)
		g.POST("", permMw("permission:create"), h.Create)
	}
}
