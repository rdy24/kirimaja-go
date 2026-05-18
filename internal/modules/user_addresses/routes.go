package user_addresses

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.RouterGroup, h *Handler, authMw gin.HandlerFunc) {
	g := r.Group("/user-addresses", authMw)
	{
		g.GET("", h.FindAll)
		g.GET("/:id", h.FindByID)
		g.POST("", h.Create)
		g.PUT("/:id", h.Update)
		g.DELETE("/:id", h.Delete)
	}
}
