package history

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.RouterGroup, h *Handler, authMw gin.HandlerFunc, permMw func(string) gin.HandlerFunc) {
	g := r.Group("/history", authMw)
	g.GET("", permMw("shipments.read"), h.FindAll)
	g.GET("/:id", h.FindByID)
}
