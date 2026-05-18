package permissions

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"kirimaja-go/internal/common/response"
)

var validate = validator.New()

type Handler struct{ svc Service }

func NewHandler(svc Service) *Handler { return &Handler{svc} }

func (h *Handler) FindAll(c *gin.Context) {
	data, err := h.svc.FindAll(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch permissions", nil)
		return
	}
	response.Success(c, http.StatusOK, "Permissions fetched", data)
}

func (h *Handler) Create(c *gin.Context) {
	var req CreatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}
	if err := validate.Struct(req); err != nil {
		response.Error(c, http.StatusUnprocessableEntity, "Validation failed", err.Error())
		return
	}
	data, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create permission", nil)
		return
	}
	response.Success(c, http.StatusCreated, "Permission created", data)
}
