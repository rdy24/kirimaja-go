package history

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"kirimaja-go/internal/common/response"
)

type Handler struct{ svc Service }

func NewHandler(svc Service) *Handler { return &Handler{svc} }

func (h *Handler) FindAll(c *gin.Context) {
	userID := c.GetUint("userID")
	roleID := c.GetUint("roleID")

	list, err := h.svc.FindAll(c.Request.Context(), userID, roleID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	response.Success(c, http.StatusOK, "History retrieved successfully", list)
}

func (h *Handler) FindByID(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	data, err := h.svc.FindByID(c.Request.Context(), uint(id))
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error(), nil)
		return
	}
	response.Success(c, http.StatusOK, "Shipment retrieved successfully", data)
}
