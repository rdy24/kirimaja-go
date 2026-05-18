package shipments

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.RouterGroup, h *Handler, authMw gin.HandlerFunc, permMw func(string) gin.HandlerFunc) {
	g := r.Group("/shipments", authMw)
	{
		g.POST("", permMw("shipments.create"), h.Create)
		g.GET("", h.FindAll)
		g.GET("/tracking/:trackingNumber", h.FindByTracking)
		g.GET("/:id", h.FindByID)
		g.GET("/:id/pdf", h.GeneratePDF)
	}
}
