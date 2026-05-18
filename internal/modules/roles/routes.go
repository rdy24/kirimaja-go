package roles

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.RouterGroup, h *Handler, authMw gin.HandlerFunc, permMw func(string) gin.HandlerFunc) {
	g := r.Group("/roles", authMw)
	{
		g.GET("", permMw("role:read"), h.FindAll)
		g.GET("/:id", permMw("role:read"), h.FindByID)
		g.POST("", permMw("role:create"), h.Create)
		g.PUT("/:id", permMw("role:update"), h.Update)
		g.DELETE("/:id", permMw("role:delete"), h.Delete)
	}
}
