package webhook

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.RouterGroup, h *Handler, mws ...gin.HandlerFunc) {
	handlers := append(append([]gin.HandlerFunc{}, mws...), h.HandleMidtrans)
	r.POST("/webhooks/midtrans", handlers...)
}
