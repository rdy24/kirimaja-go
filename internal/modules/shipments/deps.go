package shipments

import (
	"context"
	"time"

	"kirimaja-go/internal/common/midtrans"
	"kirimaja-go/internal/common/opencage"
	"kirimaja-go/internal/common/pdf"
	"kirimaja-go/internal/common/worker"
)

// The service depends on these behavioral interfaces rather than the concrete
// *opencage.Client / *midtrans.Client / etc. The concrete clients already
// satisfy them structurally (no changes there), and tests can now substitute
// fakes — previously Create/HandleWebhook could not be unit-tested without
// real network calls and a real Chrome.

// Geocoder resolves an address to coordinates. Context-aware so a slow
// geocoding provider can't pin a request goroutine forever.
type Geocoder interface {
	GeocodeContext(ctx context.Context, address string) (*opencage.Location, error)
}

// PaymentGateway is the payment provider (Midtrans Snap).
type PaymentGateway interface {
	CreateSnap(orderID string, amount int64, email string) (*midtrans.SnapResult, error)
	VerifyWebhookSignature(orderID, statusCode, grossAmount string) func(signatureKey string) bool
}

// QRGenerator renders a QR code and returns its object-storage key.
type QRGenerator interface {
	Generate(ctx context.Context, content string) (string, error)
}

// PDFRenderer renders the shipment label PDF.
type PDFRenderer interface {
	Generate(ctx context.Context, data pdf.ShipmentData) ([]byte, error)
	QRBase64(ctx context.Context, key string) string
}

// TaskQueue enqueues asynchronous jobs (emails, delayed expiry).
type TaskQueue interface {
	EnqueuePaymentNotification(p worker.PaymentNotificationPayload) error
	EnqueuePaymentSuccess(p worker.PaymentSuccessPayload) error
	EnqueuePaymentExpiry(p worker.PaymentExpiryPayload, at time.Time) error
	CancelPaymentExpiry(paymentID uint)
}
