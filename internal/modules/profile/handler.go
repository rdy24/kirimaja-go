package profile

import (
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"kirimaja-go/internal/common/response"
	"kirimaja-go/internal/common/storage"
)

var validate = validator.New()

type Handler struct {
	svc   Service
	store storage.Store
}

func NewHandler(svc Service, store storage.Store) *Handler {
	return &Handler{svc, store}
}

// withAvatarURL swaps the stored avatar object key for a presigned URL.
func (h *Handler) withAvatarURL(c *gin.Context, p *ProfileResponse) *ProfileResponse {
	if p != nil && p.Avatar != nil && *p.Avatar != "" {
		if u, err := h.store.PresignedURL(c.Request.Context(), *p.Avatar); err == nil {
			p.Avatar = &u
		} else {
			p.Avatar = nil
		}
	}
	return p
}

func (h *Handler) FindOne(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	data, err := h.svc.FindOne(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error(), nil)
		return
	}
	response.Success(c, http.StatusOK, "Profile retrieved successfully", h.withAvatarURL(c, data))
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
	data, err := h.svc.Update(c.Request.Context(), userID, req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	response.Success(c, http.StatusOK, "Profile updated successfully", h.withAvatarURL(c, data))
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

	f, err := file.Open()
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Cannot read uploaded file", nil)
		return
	}
	defer f.Close()

	key := fmt.Sprintf("photos/%d-%d%s", time.Now().UnixMilli(), userID, ext)
	if err := h.store.Put(c.Request.Context(), key, f, file.Size, file.Header.Get("Content-Type")); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to store avatar", nil)
		return
	}

	data, err := h.svc.UpdateAvatar(c.Request.Context(), userID, key)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	response.Success(c, http.StatusOK, "Avatar uploaded successfully", h.withAvatarURL(c, data))
}
