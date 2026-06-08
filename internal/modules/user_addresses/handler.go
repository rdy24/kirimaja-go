package user_addresses

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
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

func (h *Handler) presignAddr(c *gin.Context, a *UserAddressResponse) *UserAddressResponse {
	if a == nil {
		return a
	}
	ctx := c.Request.Context()
	swap := func(k *string) *string {
		if k == nil || *k == "" {
			return nil
		}
		if u, err := h.store.PresignedURL(ctx, *k); err == nil {
			return &u
		}
		return nil
	}
	a.Photo = swap(a.Photo)
	a.User.Avatar = swap(a.User.Avatar)
	return a
}

func (h *Handler) presignAddrs(c *gin.Context, list []UserAddressResponse) []UserAddressResponse {
	for i := range list {
		h.presignAddr(c, &list[i])
	}
	return list
}

// uploadOptionalPhoto stores an optional "photo" multipart field in object
// storage and returns its key, or nil if no (valid) file was sent.
func (h *Handler) uploadOptionalPhoto(c *gin.Context, userID uint) *string {
	file, err := c.FormFile("photo")
	if err != nil {
		return nil
	}
	f, err := file.Open()
	if err != nil {
		return nil
	}
	defer f.Close()
	key := fmt.Sprintf("photos/%d-%d%s", time.Now().UnixMilli(), userID, filepath.Ext(file.Filename))
	if err := h.store.Put(c.Request.Context(), key, f, file.Size, file.Header.Get("Content-Type")); err != nil {
		return nil
	}
	return &key
}

func (h *Handler) FindAll(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	data, err := h.svc.FindAll(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch addresses", nil)
		return
	}
	response.Success(c, http.StatusOK, "User addresses retrieved successfully", h.presignAddrs(c, data))
}

func (h *Handler) FindByID(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	data, err := h.svc.FindByID(c.Request.Context(), uint(id))
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error(), nil)
		return
	}
	response.Success(c, http.StatusOK, "User address retrieved successfully", h.presignAddr(c, data))
}

func (h *Handler) Create(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	var req CreateUserAddressRequest
	if err := c.ShouldBind(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}
	if err := validate.Struct(req); err != nil {
		response.Error(c, http.StatusUnprocessableEntity, "Validation failed", err.Error())
		return
	}

	photoPath := h.uploadOptionalPhoto(c, userID)

	data, err := h.svc.Create(c.Request.Context(), userID, req, photoPath)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	response.Success(c, http.StatusCreated, "User address created successfully", h.presignAddr(c, data))
}

func (h *Handler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	userID := c.MustGet("userID").(uint)

	var req UpdateUserAddressRequest
	if err := c.ShouldBind(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	photoPath := h.uploadOptionalPhoto(c, userID)

	data, err := h.svc.Update(c.Request.Context(), uint(id), req, photoPath)
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error(), nil)
		return
	}
	response.Success(c, http.StatusOK, "User address updated successfully", h.presignAddr(c, data))
}

func (h *Handler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.Delete(c.Request.Context(), uint(id)); err != nil {
		response.Error(c, http.StatusNotFound, err.Error(), nil)
		return
	}
	response.Success(c, http.StatusOK, "User address deleted successfully", nil)
}
