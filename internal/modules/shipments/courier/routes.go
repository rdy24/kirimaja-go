package courier

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.RouterGroup, h *Handler, authMw gin.HandlerFunc, permMw func(string) gin.HandlerFunc) {
	g := r.Group("/shipments/courier", authMw, permMw("delivery.read"))
	g.GET("/list", h.FindAll)
	g.GET("/pick/:trackingNumber", permMw("delivery.update"), h.PickShipment)
	g.POST("/pickup/:trackingNumber", permMw("delivery.update"), h.PickupShipment)
	g.GET("/deliver-to-branch/:trackingNumber", permMw("delivery.update"), h.DeliverToBranch)
	g.GET("/pick-from-branch/:trackingNumber", permMw("delivery.update"), h.PickShipmentFromBranch)
	g.GET("/pickup-from-branch/:trackingNumber", permMw("delivery.update"), h.PickupShipmentFromBranch)
	g.POST("/deliver-to-customer/:trackingNumber", permMw("delivery.update"), h.DeliverToCustomer)
}
