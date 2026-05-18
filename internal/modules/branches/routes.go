package branches

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.RouterGroup, h *Handler, authMw gin.HandlerFunc, permMw func(string) gin.HandlerFunc) {
	g := r.Group("/branches", authMw)
	{
		g.GET("", permMw("branch:read"), h.FindAll)
		g.GET("/:id", permMw("branch:read"), h.FindByID)
		g.POST("", permMw("branch:create"), h.Create)
		g.PUT("/:id", permMw("branch:update"), h.Update)
		g.DELETE("/:id", permMw("branch:delete"), h.Delete)
	}
}
