package profile

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.RouterGroup, h *Handler, authMw gin.HandlerFunc) {
	g := r.Group("/profile", authMw)
	{
		g.GET("", h.FindOne)
		g.PUT("", h.Update)
		g.POST("/avatar", h.UploadAvatar)
	}
}
