package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hibiken/asynq"
	"gorm.io/gorm"
	"kirimaja-go/internal/common/email"
	"kirimaja-go/models"
)

type Server struct {
	srv      *asynq.Server
	mux      *asynq.ServeMux
	emailSvc *email.Service
	db       *gorm.DB
}

func NewServer(redisAddr string, emailSvc *email.Service, db *gorm.DB) *Server {
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{Concurrency: 10, Queues: map[string]int{DefaultQueue: 1}},
	)
	mux := asynq.NewServeMux()
	s := &Server{srv: srv, mux: mux, emailSvc: emailSvc, db: db}
	mux.HandleFunc(TypePaymentNotification, s.handlePaymentNotification)
	mux.HandleFunc(TypePaymentSuccess, s.handlePaymentSuccess)
	mux.HandleFunc(TypePaymentExpiry, s.handlePaymentExpiry)
	return s
}

func (s *Server) Start() {
	go func() {
		if err := s.srv.Run(s.mux); err != nil {
			log.Fatalf("[worker] server error: %v", err)
		}
	}()
	log.Println("[worker] asynq server started")
}

func (s *Server) handlePaymentNotification(_ context.Context, t *asynq.Task) error {
	var p PaymentNotificationPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}
	if err := s.emailSvc.SendPaymentNotification(p.To, p.PaymentURL, p.ShipmentID, p.Amount, p.ExpiryDate); err != nil {
		return fmt.Errorf("send payment notification: %w", err)
	}
	log.Printf("[worker] payment notification sent to %s for shipment %d", p.To, p.ShipmentID)
	return nil
}

func (s *Server) handlePaymentSuccess(_ context.Context, t *asynq.Task) error {
	var p PaymentSuccessPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}
	if err := s.emailSvc.SendPaymentSuccess(p.To, p.ShipmentID, p.Amount, p.TrackingNumber); err != nil {
		return fmt.Errorf("send payment success: %w", err)
	}
	log.Printf("[worker] payment success email sent to %s for shipment %d", p.To, p.ShipmentID)
	return nil
}

func (s *Server) handlePaymentExpiry(_ context.Context, t *asynq.Task) error {
	var p PaymentExpiryPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	var payment models.Payment
	if err := s.db.First(&payment, p.PaymentID).Error; err != nil {
		log.Printf("[worker] payment %d not found, skipping expiry", p.PaymentID)
		return nil
	}
	if payment.Status == nil || *payment.Status != "PENDING" {
		log.Printf("[worker] payment %d is no longer pending (%v), skipping", p.PaymentID, payment.Status)
		return nil
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Payment{}).Where("id = ?", p.PaymentID).
			Update("status", "EXPIRED").Error; err != nil {
			return err
		}
		if err := tx.Model(&models.Shipment{}).Where("id = ?", p.ShipmentID).
			Update("payment_status", "EXPIRED").Error; err != nil {
			return err
		}
		desc := "Payment expired - automatic expiry"
		return tx.Create(&models.ShipmentHistory{
			ShipmentID:  p.ShipmentID,
			Status:      "EXPIRED",
			Description: &desc,
		}).Error
	})
}
