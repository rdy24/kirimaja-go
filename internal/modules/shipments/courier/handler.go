package courier

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"kirimaja-go/internal/common/response"
	"kirimaja-go/internal/modules/shipments"
)

const maxProofPhotoBytes = 5 << 20 // 5 MiB

// allowedProofPhotoTypes maps a sniffed content type to the extension we
// store it under — the client-supplied filename/extension is never trusted.
var allowedProofPhotoTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

// saveProofPhoto validates the upload (size + real content type) and writes
// it with a server-generated name. Returns the stored filename.
func (h *Handler) saveProofPhoto(c *gin.Context, file *multipart.FileHeader, userID uint) (string, error) {
	if file.Size > maxProofPhotoBytes {
		return "", fmt.Errorf("photo too large (max %d MB)", maxProofPhotoBytes>>20)
	}

	f, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("cannot read uploaded file")
	}
	head := make([]byte, 512)
	n, _ := f.Read(head)
	_ = f.Close()
	contentType := http.DetectContentType(head[:n])

	ext, ok := allowedProofPhotoTypes[contentType]
	if !ok {
		return "", fmt.Errorf("unsupported image type: %s", contentType)
	}

	dir := filepath.Join(h.publicDir, "uploads", "photos")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("cannot prepare upload directory")
	}

	filename := fmt.Sprintf("%d-%d%s", time.Now().UnixMilli(), userID, ext)
	if err := c.SaveUploadedFile(file, filepath.Join(dir, filename)); err != nil {
		return "", fmt.Errorf("failed to save photo")
	}
	return filename, nil
}

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

	filename, err := h.saveProofPhoto(c, file, userID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
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

	filename, err := h.saveProofPhoto(c, file, userID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	s, err := h.svc.DeliverToCustomer(tracking, userID, filename)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	response.Success(c, http.StatusOK, "Shipment delivered to customer successfully", s)
}
