package webhook

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.RouterGroup, h *Handler) {
	r.POST("/webhooks/midtrans", h.HandleMidtrans)
}
