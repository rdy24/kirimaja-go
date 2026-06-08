package courier

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"kirimaja-go/internal/common/response"
	"kirimaja-go/internal/common/storage"
	"kirimaja-go/internal/modules/shipments"
	"kirimaja-go/models"
)

const maxProofPhotoBytes = 5 << 20 // 5 MiB

// allowedProofPhotoTypes maps a sniffed content type to the extension we
// store it under — the client-supplied filename/extension is never trusted.
var allowedProofPhotoTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

// saveProofPhoto validates the upload (size + sniffed content type) and
// stores it in object storage under a server-generated key, which it returns.
func (h *Handler) saveProofPhoto(c *gin.Context, file *multipart.FileHeader, userID uint) (string, error) {
	if file.Size > maxProofPhotoBytes {
		return "", fmt.Errorf("photo too large (max %d MB)", maxProofPhotoBytes>>20)
	}

	f, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("cannot read uploaded file")
	}
	defer f.Close()

	head := make([]byte, 512)
	n, _ := f.Read(head)
	contentType := http.DetectContentType(head[:n])
	ext, ok := allowedProofPhotoTypes[contentType]
	if !ok {
		return "", fmt.Errorf("unsupported image type: %s", contentType)
	}
	if _, err := f.Seek(0, 0); err != nil {
		return "", fmt.Errorf("cannot rewind uploaded file")
	}

	key := fmt.Sprintf("proofs/%d-%d%s", time.Now().UnixMilli(), userID, ext)
	if err := h.store.Put(c.Request.Context(), key, f, file.Size, contentType); err != nil {
		return "", fmt.Errorf("failed to store photo")
	}
	return key, nil
}

type Handler struct {
	svc   shipments.CourierService
	store storage.Store
}

func NewHandler(svc shipments.CourierService, store storage.Store) *Handler {
	return &Handler{svc, store}
}

func (h *Handler) okShipment(c *gin.Context, msg string, s *models.Shipment) {
	resp := shipments.ToShipmentResponse(s)
	shipments.PresignShipment(c.Request.Context(), h.store, resp)
	response.Success(c, http.StatusOK, msg, resp)
}

func (h *Handler) FindAll(c *gin.Context) {
	list, err := h.svc.FindAllForCourier(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	resp := shipments.ToShipmentResponses(list)
	shipments.PresignShipments(c.Request.Context(), h.store, resp)
	response.Success(c, http.StatusOK, "Shipments retrieved successfully", resp)
}

func (h *Handler) PickShipment(c *gin.Context) {
	tracking := c.Param("trackingNumber")
	userID := c.GetUint("userID")

	s, err := h.svc.PickShipment(c.Request.Context(), tracking, userID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	h.okShipment(c, "Shipment picked up successfully", s)
}

func (h *Handler) PickupShipment(c *gin.Context) {
	tracking := c.Param("trackingNumber")
	userID := c.GetUint("userID")

	file, err := c.FormFile("photo")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Pickup proof image is required", nil)
		return
	}

	filename, err := h.saveProofPhoto(c, file, userID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	s, err := h.svc.PickupShipment(c.Request.Context(), tracking, userID, filename)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	h.okShipment(c, "Shipment pickup confirmed successfully", s)
}

func (h *Handler) DeliverToBranch(c *gin.Context) {
	tracking := c.Param("trackingNumber")
	userID := c.GetUint("userID")

	s, err := h.svc.DeliverToBranch(c.Request.Context(), tracking, userID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	h.okShipment(c, "Shipment delivered to branch successfully", s)
}

func (h *Handler) PickShipmentFromBranch(c *gin.Context) {
	tracking := c.Param("trackingNumber")
	userID := c.GetUint("userID")

	s, err := h.svc.PickShipmentFromBranch(c.Request.Context(), tracking, userID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	h.okShipment(c, "Shipment picked from branch successfully", s)
}

func (h *Handler) PickupShipmentFromBranch(c *gin.Context) {
	tracking := c.Param("trackingNumber")
	userID := c.GetUint("userID")

	s, err := h.svc.PickupShipmentFromBranch(c.Request.Context(), tracking, userID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	h.okShipment(c, "Shipment picked up from branch successfully", s)
}

func (h *Handler) DeliverToCustomer(c *gin.Context) {
	tracking := c.Param("trackingNumber")
	userID := c.GetUint("userID")

	file, err := c.FormFile("photo")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Receipt proof image is required", nil)
		return
	}

	filename, err := h.saveProofPhoto(c, file, userID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	s, err := h.svc.DeliverToCustomer(c.Request.Context(), tracking, userID, filename)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	h.okShipment(c, "Shipment delivered to customer successfully", s)
}
