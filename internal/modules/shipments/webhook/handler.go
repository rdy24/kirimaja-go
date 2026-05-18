package webhook

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"kirimaja-go/internal/common/response"
	"kirimaja-go/internal/modules/shipments"
)

type Handler struct {
	svc shipments.WebhookService
}

func NewHandler(svc shipments.WebhookService) *Handler {
	return &Handler{svc}
}

func (h *Handler) HandleMidtrans(c *gin.Context) {
	var payload MidtransWebhookPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid payload", err.Error())
		return
	}

	err := h.svc.HandleWebhook(shipments.WebhookPayload{
		TransactionID:     payload.TransactionID,
		TransactionStatus: payload.TransactionStatus,
		OrderID:           payload.OrderID,
		GrossAmount:       payload.GrossAmount,
		StatusCode:        payload.StatusCode,
		SignatureKey:      payload.SignatureKey,
		PaymentType:       payload.PaymentType,
	})
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	response.Success(c, http.StatusOK, "Webhook processed successfully", nil)
}
