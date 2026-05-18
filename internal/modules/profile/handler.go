package profile

import (
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"kirimaja-go/internal/common/response"
)

var validate = validator.New()

type Handler struct {
	svc       Service
	publicDir string
}

func NewHandler(svc Service, publicDir string) *Handler {
	return &Handler{svc, publicDir}
}

func (h *Handler) FindOne(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	data, err := h.svc.FindOne(userID)
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error(), nil)
		return
	}
	response.Success(c, http.StatusOK, "Profile retrieved successfully", data)
}

func (h *Handler) Update(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}
	if err := validate.Struct(req); err != nil {
		response.Error(c, http.StatusUnprocessableEntity, "Validation failed", err.Error())
		return
	}
	data, err := h.svc.Update(userID, req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	response.Success(c, http.StatusOK, "Profile updated successfully", data)
}

func (h *Handler) UploadAvatar(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	file, err := c.FormFile("avatar")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Avatar file is required", err.Error())
		return
	}

	ext := filepath.Ext(file.Filename)
	allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".avif": true}
	if !allowed[ext] {
		response.Error(c, http.StatusBadRequest, "Only image files are allowed (jpg, jpeg, png, gif, avif)", nil)
		return
	}

	filename := fmt.Sprintf("%d-%d%s", time.Now().UnixMilli(), userID, ext)
	savePath := filepath.Join(h.publicDir, "uploads", "photos", filename)
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to save avatar", nil)
		return
	}

	avatarPath := "/uploads/photos/" + filename
	data, err := h.svc.UpdateAvatar(userID, avatarPath)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	response.Success(c, http.StatusOK, "Avatar uploaded successfully", data)
}
