package worker

import "time"

const (
	TypePaymentNotification = "email:payment-notification"
	TypePaymentSuccess      = "email:payment-success"
	TypePaymentExpiry       = "payment:expiry"

	DefaultQueue = "default"
)

type PaymentNotificationPayload struct {
	To         string
	PaymentURL string
	ShipmentID uint
	Amount     float64
	ExpiryDate time.Time
}

type PaymentSuccessPayload struct {
	To             string
	ShipmentID     uint
	Amount         float64
	TrackingNumber string
}

type PaymentExpiryPayload struct {
	PaymentID  uint
	ShipmentID uint
	ExternalID string
}
