package branch

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"kirimaja-go/internal/common/response"
	"kirimaja-go/internal/modules/shipments"
)

type Handler struct {
	svc shipments.BranchService
}

func NewHandler(svc shipments.BranchService) *Handler {
	return &Handler{svc}
}

func (h *Handler) FindAll(c *gin.Context) {
	userID := c.GetUint("userID")
	roleID := c.GetUint("roleID")

	logs, err := h.svc.FindAllBranchLogs(c.Request.Context(), userID, roleID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	response.Success(c, http.StatusOK, "Shipment logs retrieved successfully", logs)
}

func (h *Handler) Scan(c *gin.Context) {
	var req shipments.ScanShipmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid payload", err.Error())
		return
	}

	userID := c.GetUint("userID")
	log, err := h.svc.ScanShipment(c.Request.Context(), req, userID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	response.Success(c, http.StatusOK, "Shipment scanned successfully", log)
}
