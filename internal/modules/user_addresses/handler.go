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
)

var validate = validator.New()

type Handler struct {
	svc       Service
	publicDir string
}

func NewHandler(svc Service, publicDir string) *Handler {
	return &Handler{svc, publicDir}
}

func (h *Handler) FindAll(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	data, err := h.svc.FindAll(userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch addresses", nil)
		return
	}
	response.Success(c, http.StatusOK, "User addresses retrieved successfully", data)
}

func (h *Handler) FindByID(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	data, err := h.svc.FindByID(uint(id))
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error(), nil)
		return
	}
	response.Success(c, http.StatusOK, "User address retrieved successfully", data)
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

	var photoPath *string
	if file, err := c.FormFile("photo"); err == nil {
		ext := filepath.Ext(file.Filename)
		filename := fmt.Sprintf("%d-%d%s", time.Now().UnixMilli(), userID, ext)
		savePath := filepath.Join(h.publicDir, "uploads", "photos", filename)
		if err := c.SaveUploadedFile(file, savePath); err == nil {
			p := "/uploads/photos/" + filename
			photoPath = &p
		}
	}

	data, err := h.svc.Create(userID, req, photoPath)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	response.Success(c, http.StatusCreated, "User address created successfully", data)
}

func (h *Handler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	userID := c.MustGet("userID").(uint)

	var req UpdateUserAddressRequest
	if err := c.ShouldBind(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	var photoPath *string
	if file, err := c.FormFile("photo"); err == nil {
		ext := filepath.Ext(file.Filename)
		filename := fmt.Sprintf("%d-%d%s", time.Now().UnixMilli(), userID, ext)
		savePath := filepath.Join(h.publicDir, "uploads", "photos", filename)
		if err := c.SaveUploadedFile(file, savePath); err == nil {
			p := "/uploads/photos/" + filename
			photoPath = &p
		}
	}

	data, err := h.svc.Update(uint(id), req, photoPath)
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error(), nil)
		return
	}
	response.Success(c, http.StatusOK, "User address updated successfully", data)
}

func (h *Handler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.Delete(uint(id)); err != nil {
		response.Error(c, http.StatusNotFound, err.Error(), nil)
		return
	}
	response.Success(c, http.StatusOK, "User address deleted successfully", nil)
}
