package shipments

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"kirimaja-go/internal/common/response"
	"kirimaja-go/internal/common/storage"
)

var validate = validator.New()

type Handler struct {
	svc   ShipmentService
	store storage.Store
}

func NewHandler(svc ShipmentService, store storage.Store) *Handler {
	return &Handler{svc, store}
}

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
	data, err := h.svc.Create(c.Request.Context(), userID, req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	resp := ToShipmentResponse(data)
	PresignShipment(c.Request.Context(), h.store, resp)
	response.Success(c, http.StatusCreated, "Shipment created successfully", resp)
}

func (h *Handler) FindAll(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	data, err := h.svc.FindAll(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch shipments", nil)
		return
	}
	resp := ToShipmentResponses(data)
	PresignShipments(c.Request.Context(), h.store, resp)
	response.Success(c, http.StatusOK, "Shipments retrieved successfully", resp)
}

func (h *Handler) FindByID(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	userID := c.MustGet("userID").(uint)
	roleID := c.GetUint("roleID")
	data, err := h.svc.FindByID(c.Request.Context(), uint(id), userID, roleID)
	if err != nil {
		respondShipmentErr(c, err)
		return
	}
	resp := ToShipmentResponse(data)
	PresignShipment(c.Request.Context(), h.store, resp)
	response.Success(c, http.StatusOK, "Shipment retrieved successfully", resp)
}

func (h *Handler) FindByTracking(c *gin.Context) {
	tracking := c.Param("trackingNumber")
	userID := c.MustGet("userID").(uint)
	roleID := c.GetUint("roleID")
	data, err := h.svc.FindByTrackingNumber(c.Request.Context(), tracking, userID, roleID)
	if err != nil {
		respondShipmentErr(c, err)
		return
	}
	resp := ToShipmentResponse(data)
	PresignShipment(c.Request.Context(), h.store, resp)
	response.Success(c, http.StatusOK, "Shipment retrieved successfully", resp)
}

// respondShipmentErr maps a forbidden access to 403 (not 404, which would
// otherwise leak the same status whether or not the shipment exists — but
// here we deliberately avoid revealing existence to non-owners).
func respondShipmentErr(c *gin.Context, err error) {
	if errors.Is(err, ErrForbidden) {
		response.Error(c, http.StatusForbidden, err.Error(), nil)
		return
	}
	response.Error(c, http.StatusNotFound, err.Error(), nil)
}

func (h *Handler) GeneratePDF(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	userID := c.MustGet("userID").(uint)
	roleID := c.GetUint("roleID")
	pdfBytes, err := h.svc.GeneratePDFByID(c.Request.Context(), uint(id), userID, roleID)
	if err != nil {
		if errors.Is(err, ErrForbidden) {
			response.Error(c, http.StatusForbidden, err.Error(), nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=shipment-%d.pdf", id))
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}
