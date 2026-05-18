package employee_branches

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.RouterGroup, h *Handler, authMw gin.HandlerFunc, permMw func(string) gin.HandlerFunc) {
	g := r.Group("/employee-branches", authMw)
	{
		g.GET("", permMw("employee-branch:read"), h.FindAll)
		g.GET("/:id", permMw("employee-branch:read"), h.FindByID)
		g.POST("", permMw("employee-branch:create"), h.Create)
		g.PUT("/:id", permMw("employee-branch:update"), h.Update)
		g.DELETE("/:id", permMw("employee-branch:delete"), h.Delete)
	}
}
