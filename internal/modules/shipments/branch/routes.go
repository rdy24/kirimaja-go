package branch

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.RouterGroup, h *Handler, authMw gin.HandlerFunc, permMw func(string) gin.HandlerFunc) {
	g := r.Group("/shipments/branch", authMw)
	g.GET("/logs", permMw("shipment-branch.read"), h.FindAll)
	g.POST("/scan", h.Scan)
}
