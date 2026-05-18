package shipments

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"kirimaja-go/internal/common/response"
)

var validate = validator.New()

type Handler struct{ svc Service }

func NewHandler(svc Service) *Handler { return &Handler{svc} }

func (h *Handler) Create(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	var req CreateShipmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}
	if err := validate.Struct(req); err != nil {
		response.Error(c, http.StatusUnprocessableEntity, "Validation failed", err.Error())
		return
	}
	data, err := h.svc.Create(userID, req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	response.Success(c, http.StatusCreated, "Shipment created successfully", data)
}

func (h *Handler) FindAll(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	data, err := h.svc.FindAll(userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch shipments", nil)
		return
	}
	response.Success(c, http.StatusOK, "Shipments retrieved successfully", data)
}

func (h *Handler) FindByID(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	data, err := h.svc.FindByID(uint(id))
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error(), nil)
		return
	}
	response.Success(c, http.StatusOK, "Shipment retrieved successfully", data)
}

func (h *Handler) FindByTracking(c *gin.Context) {
	tracking := c.Param("trackingNumber")
	data, err := h.svc.FindByTrackingNumber(tracking)
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error(), nil)
		return
	}
	response.Success(c, http.StatusOK, "Shipment retrieved successfully", data)
}

func (h *Handler) GeneratePDF(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	pdfBytes, err := h.svc.GeneratePDFByID(uint(id))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=shipment-%d.pdf", id))
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}
