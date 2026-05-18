package courier

import (
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"kirimaja-go/internal/common/response"
	"kirimaja-go/internal/modules/shipments"
)

type Handler struct {
	svc       shipments.Service
	publicDir string
}

func NewHandler(svc shipments.Service, publicDir string) *Handler {
	return &Handler{svc, publicDir}
}

func (h *Handler) FindAll(c *gin.Context) {
	list, err := h.svc.FindAllForCourier()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	response.Success(c, http.StatusOK, "Shipments retrieved successfully", list)
}

func (h *Handler) PickShipment(c *gin.Context) {
	tracking := c.Param("trackingNumber")
	userID := c.GetUint("userID")

	s, err := h.svc.PickShipment(tracking, userID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	response.Success(c, http.StatusOK, "Shipment picked up successfully", s)
}

func (h *Handler) PickupShipment(c *gin.Context) {
	tracking := c.Param("trackingNumber")
	userID := c.GetUint("userID")

	file, err := c.FormFile("photo")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Pickup proof image is required", nil)
		return
	}

	filename := fmt.Sprintf("%d-%d%s", time.Now().UnixMilli(), userID, filepath.Ext(file.Filename))
	savePath := filepath.Join(h.publicDir, "uploads", "photos", filename)
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to save photo", nil)
		return
	}

	s, err := h.svc.PickupShipment(tracking, userID, filename)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	response.Success(c, http.StatusOK, "Shipment pickup confirmed successfully", s)
}

func (h *Handler) DeliverToBranch(c *gin.Context) {
	tracking := c.Param("trackingNumber")
	userID := c.GetUint("userID")

	s, err := h.svc.DeliverToBranch(tracking, userID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	response.Success(c, http.StatusOK, "Shipment delivered to branch successfully", s)
}

func (h *Handler) PickShipmentFromBranch(c *gin.Context) {
	tracking := c.Param("trackingNumber")
	userID := c.GetUint("userID")

	s, err := h.svc.PickShipmentFromBranch(tracking, userID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	response.Success(c, http.StatusOK, "Shipment picked from branch successfully", s)
}

func (h *Handler) PickupShipmentFromBranch(c *gin.Context) {
	tracking := c.Param("trackingNumber")
	userID := c.GetUint("userID")

	s, err := h.svc.PickupShipmentFromBranch(tracking, userID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	response.Success(c, http.StatusOK, "Shipment picked up from branch successfully", s)
}

func (h *Handler) DeliverToCustomer(c *gin.Context) {
	tracking := c.Param("trackingNumber")
	userID := c.GetUint("userID")

	file, err := c.FormFile("photo")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Receipt proof image is required", nil)
		return
	}

	filename := fmt.Sprintf("%d-%d%s", time.Now().UnixMilli(), userID, filepath.Ext(file.Filename))
	savePath := filepath.Join(h.publicDir, "uploads", "photos", filename)
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to save photo", nil)
		return
	}

	s, err := h.svc.DeliverToCustomer(tracking, userID, filename)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	response.Success(c, http.StatusOK, "Shipment delivered to customer successfully", s)
}
