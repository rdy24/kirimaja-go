package auth

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"kirimaja-go/internal/common/response"
)

var validate = validator.New()

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc}
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}
	if err := validate.Struct(req); err != nil {
		response.Error(c, http.StatusUnprocessableEntity, "Validation failed", err.Error())
		return
	}

	result, err := h.svc.Login(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			response.Error(c, http.StatusUnauthorized, err.Error(), nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "Login failed", nil)
		return
	}

	response.Success(c, http.StatusOK, "Login successful", result)
}

func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}
	if err := validate.Struct(req); err != nil {
		response.Error(c, http.StatusUnprocessableEntity, "Validation failed", err.Error())
		return
	}

	result, err := h.svc.Register(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, ErrEmailRegistered) {
			response.Error(c, http.StatusConflict, err.Error(), nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "Register failed", nil)
		return
	}

	response.Success(c, http.StatusCreated, "Register successful", result)
}
