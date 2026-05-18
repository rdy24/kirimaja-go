package shipments

import (
	"errors"
	"fmt"
	"math"
	"time"

	"gorm.io/gorm"
	"kirimaja-go/internal/common/midtrans"
	"kirimaja-go/internal/common/opencage"
	"kirimaja-go/internal/common/pdf"
	"kirimaja-go/internal/common/qrcode"
	"kirimaja-go/internal/common/worker"
	"kirimaja-go/models"
)

// Payment statuses
const (
	StatusPending  = "PENDING"
	StatusPaid     = "PAID"
	StatusSettled  = "SETTLED"
	StatusExpired  = "EXPIRED"
	StatusFailed   = "FAILED"
)

// Delivery statuses
const (
	StatusReadyToPickup          = "READY_TO_PICKUP"
	StatusWaitingPickup          = "WAITING_PICKUP"
	StatusPickedUp               = "PICKED_UP"
	StatusInTransit              = "IN_TRANSIT"
	StatusArrivedAtBranch        = "ARRIVED_AT_BRANCH"
	StatusDepartedFromBranch     = "DEPARTED_FROM_BRANCH"
	StatusReadyToPickupAtBranch  = "READY_TO_PICKUP_AT_BRANCH"
	StatusReadyToDeliver         = "READY_TO_DELIVER"
	StatusOnTheWayToAddress      = "ON_THE_WAY_TO_ADDRESS"
	StatusOnTheWay               = "ON_THE_WAY"
	StatusDelivered              = "DELIVERED"
)

// ErrForbidden is returned when an authenticated user tries to read a
// shipment they do not own (and is not a privileged role).
var ErrForbidden = errors.New("you do not have access to this shipment")

type WebhookPayload struct {
	TransactionID     string
	TransactionStatus string
	OrderID           string
	GrossAmount       string
	StatusCode        string
	SignatureKey      string
	PaymentType       string
}

type ScanShipmentRequest struct {
	TrackingNumber  string `json:"tracking_number" binding:"required"`
	Type            string `json:"type" binding:"required,oneof=IN OUT"`
	IsReadyToPickup bool   `json:"is_ready_to_pickup"`
}

type Service interface {
	Create(userID uint, req CreateShipmentRequest) (*models.Shipment, error)
	FindAll(userID uint) ([]models.Shipment, error)
	FindByID(id, userID, roleID uint) (*models.Shipment, error)
	FindByTrackingNumber(tracking string, userID, roleID uint) (*models.Shipment, error)
	HandleWebhook(payload WebhookPayload) error

	// Branch
	FindAllBranchLogs(userID, roleID uint) ([]models.ShipmentBranchLog, error)
	ScanShipment(req ScanShipmentRequest, userID uint) (*models.ShipmentBranchLog, error)

	// Courier
	FindAllForCourier() ([]models.Shipment, error)
	PickShipment(trackingNumber string, userID uint) (*models.Shipment, error)
	PickupShipment(trackingNumber string, userID uint, photoFilename string) (*models.Shipment, error)
	DeliverToBranch(trackingNumber string, userID uint) (*models.Shipment, error)
	PickShipmentFromBranch(trackingNumber string, userID uint) (*models.Shipment, error)
	PickupShipmentFromBranch(trackingNumber string, userID uint) (*models.Shipment, error)
	DeliverToCustomer(trackingNumber string, userID uint, photoFilename string) (*models.Shipment, error)

	// PDF
	GeneratePDFByID(id, userID, roleID uint) ([]byte, error)
}

type service struct {
	repo     Repository
	geocli   *opencage.Client
	midtrans *midtrans.Client
	qrSvc    *qrcode.Service
	worker   *worker.Client
	pdfSvc   *pdf.Service
}

func NewService(repo Repository, geo *opencage.Client, mt *midtrans.Client, qr *qrcode.Service, wk *worker.Client, pdfSvc *pdf.Service) Service {
	return &service{repo, geo, mt, qr, wk, pdfSvc}
}

func (s *service) Create(userID uint, req CreateShipmentRequest) (*models.Shipment, error) {
	loc, err := s.geocli.Geocode(req.DestinationAddress)
	if err != nil {
		return nil, fmt.Errorf("geocode gagal: %w", err)
	}

	addr, err := s.repo.FindPickupAddress(req.PickupAddressID)
	if err != nil || addr == nil {
		return nil, errors.New("pickup address tidak ditemukan")
	}
	if addr.Latitude == nil || addr.Longitude == nil {
		return nil, errors.New("pickup address belum memiliki koordinat")
	}

	distKm := opencage.HaversineKm(*addr.Latitude, *addr.Longitude, loc.Lat, loc.Lng)
	cost := calculateShipmentCost(distKm, req.Weight, req.DeliveryType)

	// Call the external payment gateway BEFORE any DB write. If it fails we
	// persisted nothing — no orphan shipment without a payment row. orderID
	// no longer embeds shipment.ID (which doesn't exist yet); UnixNano keeps
	// it unique and it's only ever looked up via external_id.
	orderID := fmt.Sprintf("INV-%d", time.Now().UnixNano())
	snap, err := s.midtrans.CreateSnap(orderID, int64(cost.TotalPrice), addr.User.Email)
	if err != nil {
		return nil, fmt.Errorf("midtrans error: %w", err)
	}

	expiryDate := time.Now().Add(24 * time.Hour)
	var shipment models.Shipment
	var payment models.Payment
	err = s.repo.Transaction(func(tx *gorm.DB) error {
		shipment = models.Shipment{
			PaymentStatus: StatusPending,
			Distance:      &distKm,
			Price:         &cost.TotalPrice,
		}
		if err := s.repo.CreateShipment(tx, &shipment); err != nil {
			return err
		}
		detail := models.ShipmentDetail{
			ShipmentID:           shipment.ID,
			UserID:               addr.UserID,
			PickupAddressID:      req.PickupAddressID,
			DestinationAddress:   req.DestinationAddress,
			DestinationLatitude:  &loc.Lat,
			DestinationLongitude: &loc.Lng,
			RecipientName:        req.RecipientName,
			RecipientPhone:       req.RecipientPhone,
			Weight:               &req.Weight,
			PackageType:          req.PackageType,
			DeliveryType:         req.DeliveryType,
			BasePrice:            &cost.BasePrice,
			WeightPrice:          &cost.WeightPrice,
			DistancePrice:        &cost.DistancePrice,
		}
		if err := s.repo.CreateShipmentDetail(tx, &detail); err != nil {
			return err
		}
		payment = models.Payment{
			ShipmentID: shipment.ID,
			ExternalID: &orderID,
			Status:     strPtr(StatusPending),
			InvoiceUrl: &snap.RedirectURL,
			ExpiryDate: &expiryDate,
		}
		if err := s.repo.CreatePayment(tx, &payment); err != nil {
			return err
		}
		desc := fmt.Sprintf("Shipment created with total price %.0f", cost.TotalPrice)
		return s.repo.CreateHistory(tx, &models.ShipmentHistory{
			ShipmentID:  shipment.ID,
			Status:      StatusPending,
			Description: &desc,
		})
	})
	if err != nil {
		return nil, err
	}

	if s.worker != nil {
		_ = s.worker.EnqueuePaymentNotification(worker.PaymentNotificationPayload{
			To:         addr.User.Email,
			PaymentURL: snap.RedirectURL,
			ShipmentID: shipment.ID,
			Amount:     cost.TotalPrice,
			ExpiryDate: expiryDate,
		})
		_ = s.worker.EnqueuePaymentExpiry(worker.PaymentExpiryPayload{
			PaymentID:  payment.ID,
			ShipmentID: shipment.ID,
			ExternalID: orderID,
		}, expiryDate)
	}

	return &shipment, nil
}

func (s *service) HandleWebhook(payload WebhookPayload) error {
	// 1. Verify Midtrans signature
	verify := s.midtrans.VerifyWebhookSignature(payload.OrderID, payload.StatusCode, payload.GrossAmount)
	if !verify(payload.SignatureKey) {
		return errors.New("invalid webhook signature")
	}

	// 2. Find payment by order_id (stored as external_id)
	payment, err := s.repo.FindPaymentByExternalID(payload.OrderID)
	if err != nil || payment == nil {
		return fmt.Errorf("payment with order_id %s not found", payload.OrderID)
	}

	// 3. Map Midtrans status → internal
	internalStatus := mapMidtransStatus(payload.TransactionStatus)

	// Idempotency: Midtrans retries notifications. Once a payment has reached
	// a terminal state, replays must not regenerate the QR, overwrite the
	// tracking number, re-insert history, or re-enqueue the success email.
	current := ""
	if payment.Status != nil {
		current = *payment.Status
	}
	if current == StatusPaid || current == StatusSettled {
		return nil
	}
	if (current == StatusExpired || current == StatusFailed) &&
		(internalStatus == StatusExpired || internalStatus == StatusFailed) {
		return nil
	}

	err = s.repo.Transaction(func(tx *gorm.DB) error {
		// 4. Update payment
		if err := s.repo.UpdatePayment(tx, payment.ID, map[string]any{
			"status":         internalStatus,
			"payment_method": payload.PaymentType,
		}); err != nil {
			return err
		}

		if internalStatus == StatusPaid || internalStatus == StatusSettled {
			// 5. Generate tracking number + QR code
			trackingNumber := fmt.Sprintf("KA%s", payload.TransactionID)

			qrPath, err := s.qrSvc.Generate(trackingNumber)
			if err != nil {
				return fmt.Errorf("qr code generation failed: %w", err)
			}

			// 6. Update shipment
			if err := s.repo.UpdateShipment(tx, payment.ShipmentID, map[string]any{
				"tracking_number":  trackingNumber,
				"delivery_status":  StatusReadyToPickup,
				"payment_status":   internalStatus,
				"qr_code_image":    qrPath,
			}); err != nil {
				return err
			}

			// 7. Add history
			userID := payment.Shipment.ShipmentDetail.UserID
			desc := fmt.Sprintf("Payment %s. Tracking number: %s", internalStatus, trackingNumber)
			if err := s.repo.CreateHistory(tx, &models.ShipmentHistory{
				ShipmentID:  payment.ShipmentID,
				UserID:      &userID,
				Status:      StatusReadyToPickup,
				Description: &desc,
			}); err != nil {
				return err
			}

			if s.worker != nil {
				s.worker.CancelPaymentExpiry(payment.ID)
				var price float64
				if payment.Shipment.Price != nil {
					price = *payment.Shipment.Price
				}
				_ = s.worker.EnqueuePaymentSuccess(worker.PaymentSuccessPayload{
					To:             payment.Shipment.ShipmentDetail.User.Email,
					ShipmentID:     payment.ShipmentID,
					Amount:         price,
					TrackingNumber: trackingNumber,
				})
			}

		} else if internalStatus == StatusExpired || internalStatus == StatusFailed {
			if err := s.repo.UpdateShipment(tx, payment.ShipmentID, map[string]any{
				"payment_status": internalStatus,
			}); err != nil {
				return err
			}
			desc := fmt.Sprintf("Payment %s", internalStatus)
			if err := s.repo.CreateHistory(tx, &models.ShipmentHistory{
				ShipmentID:  payment.ShipmentID,
				Status:      internalStatus,
				Description: &desc,
			}); err != nil {
				return err
			}
		}

		return nil
	})

	return err
}

func (s *service) FindAll(userID uint) ([]models.Shipment, error) {
	return s.repo.FindAll(userID)
}

// authorizeShipmentAccess enforces object-level authorization: privileged
// roles see everything, everyone else only their own shipments. Without this
// any authenticated user could read any shipment by enumerating IDs.
func authorizeShipmentAccess(shipment *models.Shipment, userID, roleID uint) error {
	if roleID == superAdminRoleID {
		return nil
	}
	if shipment.ShipmentDetail == nil || shipment.ShipmentDetail.UserID != userID {
		return ErrForbidden
	}
	return nil
}

func (s *service) FindByID(id, userID, roleID uint) (*models.Shipment, error) {
	shipment, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if shipment == nil {
		return nil, fmt.Errorf("shipment with ID %d not found", id)
	}
	if err := authorizeShipmentAccess(shipment, userID, roleID); err != nil {
		return nil, err
	}
	return shipment, nil
}

func (s *service) FindByTrackingNumber(tracking string, userID, roleID uint) (*models.Shipment, error) {
	shipment, err := s.repo.FindByTrackingNumber(tracking)
	if err != nil {
		return nil, err
	}
	if shipment == nil {
		return nil, fmt.Errorf("shipment with tracking number %s not found", tracking)
	}
	if err := authorizeShipmentAccess(shipment, userID, roleID); err != nil {
		return nil, err
	}
	return shipment, nil
}

func mapMidtransStatus(transactionStatus string) string {
	switch transactionStatus {
	case "settlement", "capture":
		return StatusPaid
	case "pending":
		return StatusPending
	case "expire":
		return StatusExpired
	case "cancel", "deny":
		return StatusFailed
	default:
		return transactionStatus
	}
}

func calculateShipmentCost(distance, weight float64, deliveryType string) ShipmentCost {
	baseRates := map[string]float64{
		"same_day": 15000, "next_day": 10000, "regular": 5000,
	}
	weightRates := map[string]float64{
		"same_day": 1000, "next_day": 800, "regular": 500,
	}
	type tierRate struct{ tier1, tier2, tier3 float64 }
	distanceRates := map[string]tierRate{
		"same_day": {8000, 12000, 15000},
		"next_day": {6000, 9000, 12000},
		"regular":  {4000, 6000, 8000},
	}

	dt := deliveryType
	if _, ok := baseRates[dt]; !ok {
		dt = "regular"
	}

	basePrice := baseRates[dt]
	weightKg := math.Ceil(weight / 1000)
	weightPrice := weightKg * weightRates[dt]

	tr := distanceRates[dt]
	var distancePrice float64
	if distance <= 50 {
		distancePrice = tr.tier1
	} else if distance <= 100 {
		distancePrice = tr.tier1 + tr.tier2
	} else {
		extra := math.Ceil((distance - 100) / 100)
		distancePrice = tr.tier3 + extra*tr.tier3
	}

	total := math.Max(basePrice+weightPrice+distancePrice, 10000)
	return ShipmentCost{
		TotalPrice: total, BasePrice: basePrice,
		WeightPrice: weightPrice, DistancePrice: distancePrice,
	}
}

func strPtr(s string) *string { return &s }

// superAdminRoleID is the role_id for SUPER_ADMIN in the roles table.
const superAdminRoleID uint = 1

func (s *service) FindAllBranchLogs(userID, roleID uint) ([]models.ShipmentBranchLog, error) {
	if roleID == superAdminRoleID {
		return s.repo.FindAllBranchLogs(nil)
	}
	eb, err := s.repo.FindEmployeeBranch(userID)
	if err != nil || eb == nil {
		return nil, errors.New("user branch not found")
	}
	return s.repo.FindAllBranchLogs(&eb.BranchID)
}

func (s *service) ScanShipment(req ScanShipmentRequest, userID uint) (*models.ShipmentBranchLog, error) {
	eb, err := s.repo.FindEmployeeBranch(userID)
	if err != nil || eb == nil {
		return nil, errors.New("user branch not found")
	}

	shipment, err := s.repo.FindByTrackingNumber(req.TrackingNumber)
	if err != nil || shipment == nil {
		return nil, errors.New("shipment not found")
	}

	validStatuses := map[string]bool{
		StatusInTransit: true, StatusArrivedAtBranch: true, StatusDepartedFromBranch: true,
	}
	ds := ""
	if shipment.DeliveryStatus != nil {
		ds = *shipment.DeliveryStatus
	}
	if !validStatuses[ds] {
		return nil, fmt.Errorf("shipment status %s is not valid for scanning", ds)
	}

	if req.Type == "OUT" {
		last, err := s.repo.FindLastBranchLogIn(req.TrackingNumber, eb.BranchID)
		if err != nil || last == nil {
			return nil, errors.New("no IN scan found for this shipment at this branch")
		}
	}

	newStatus := scanStatus(req.Type, req.IsReadyToPickup)
	desc := scanDescription(req.Type, eb.Branch.Name)

	var branchLog *models.ShipmentBranchLog
	err = s.repo.Transaction(func(tx *gorm.DB) error {
		log := &models.ShipmentBranchLog{
			ShipmentID:      shipment.ID,
			BranchID:        eb.BranchID,
			Type:            req.Type,
			Description:     &desc,
			Status:          newStatus,
			ScannedByUserID: userID,
			TrackingNumber:  req.TrackingNumber,
		}
		result, err := s.repo.CreateBranchLog(tx, log)
		if err != nil {
			return err
		}
		branchLog = result

		// The scanning branch takes ownership so its last-mile courier can
		// later pick the shipment up, and couriers from other branches can't.
		if err := s.repo.UpdateShipment(tx, shipment.ID, map[string]any{
			"delivery_status":   newStatus,
			"current_branch_id": eb.BranchID,
		}); err != nil {
			return err
		}
		return s.repo.CreateHistory(tx, &models.ShipmentHistory{
			ShipmentID:  shipment.ID,
			UserID:      &userID,
			BranchID:    &eb.BranchID,
			Status:      newStatus,
			Description: &desc,
		})
	})
	return branchLog, err
}

func scanStatus(scanType string, isReadyToPickup bool) string {
	if isReadyToPickup {
		return StatusReadyToPickupAtBranch
	}
	if scanType == "IN" {
		return StatusArrivedAtBranch
	}
	return StatusDepartedFromBranch
}

func scanDescription(scanType, branchName string) string {
	if scanType == "IN" {
		return fmt.Sprintf("Shipment arrived at %s", branchName)
	}
	return fmt.Sprintf("Shipment departed from %s", branchName)
}

func (s *service) FindAllForCourier() ([]models.Shipment, error) {
	return s.repo.FindAllForCourier()
}

func (s *service) courierShipmentAndBranch(trackingNumber string, userID uint) (*models.Shipment, *models.EmployeeBranch, error) {
	shipment, err := s.repo.FindByTrackingNumber(trackingNumber)
	if err != nil || shipment == nil {
		return nil, nil, fmt.Errorf("shipment with tracking number %s not found", trackingNumber)
	}
	eb, err := s.repo.FindEmployeeBranch(userID)
	if err != nil || eb == nil {
		return nil, nil, fmt.Errorf("user %d not found in any branch", userID)
	}
	return shipment, eb, nil
}

// assertCourierBranch enforces that the shipment has been claimed by the
// courier's own branch. Without it any courier could process any shipment
// in the system just by knowing its tracking number.
func assertCourierBranch(shipment *models.Shipment, eb *models.EmployeeBranch) error {
	if shipment.CurrentBranchID == nil || *shipment.CurrentBranchID != eb.BranchID {
		return ErrForbidden
	}
	return nil
}

func (s *service) updateCourierStatus(trackingNumber string, userID uint, newStatus, desc string) (*models.Shipment, error) {
	shipment, eb, err := s.courierShipmentAndBranch(trackingNumber, userID)
	if err != nil {
		return nil, err
	}
	if err := assertCourierBranch(shipment, eb); err != nil {
		return nil, err
	}
	err = s.repo.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.UpdateShipment(tx, shipment.ID, map[string]any{"delivery_status": newStatus}); err != nil {
			return err
		}
		return s.repo.CreateHistory(tx, &models.ShipmentHistory{
			ShipmentID:  shipment.ID,
			UserID:      &userID,
			BranchID:    &eb.BranchID,
			Status:      newStatus,
			Description: &desc,
		})
	})
	if err != nil {
		return nil, err
	}
	shipment.DeliveryStatus = &newStatus
	return shipment, nil
}

// PickShipment is the first courier touch. It "claims" an unclaimed shipment
// to the courier's branch; a shipment already claimed by another branch is
// rejected. From here on every other courier action requires the same branch.
func (s *service) PickShipment(trackingNumber string, userID uint) (*models.Shipment, error) {
	shipment, eb, err := s.courierShipmentAndBranch(trackingNumber, userID)
	if err != nil {
		return nil, err
	}
	if shipment.CurrentBranchID != nil && *shipment.CurrentBranchID != eb.BranchID {
		return nil, ErrForbidden
	}

	desc := fmt.Sprintf("Shipment picked up by user with ID %d", userID)
	err = s.repo.Transaction(func(tx *gorm.DB) error {
		upd := map[string]any{"delivery_status": StatusWaitingPickup}
		if shipment.CurrentBranchID == nil {
			upd["current_branch_id"] = eb.BranchID // claim
		}
		if err := s.repo.UpdateShipment(tx, shipment.ID, upd); err != nil {
			return err
		}
		return s.repo.CreateHistory(tx, &models.ShipmentHistory{
			ShipmentID:  shipment.ID,
			UserID:      &userID,
			BranchID:    &eb.BranchID,
			Status:      StatusWaitingPickup,
			Description: &desc,
		})
	})
	if err != nil {
		return nil, err
	}
	st := StatusWaitingPickup
	shipment.DeliveryStatus = &st
	bid := eb.BranchID
	shipment.CurrentBranchID = &bid
	return shipment, nil
}

func (s *service) PickupShipment(trackingNumber string, userID uint, photoFilename string) (*models.Shipment, error) {
	shipment, eb, err := s.courierShipmentAndBranch(trackingNumber, userID)
	if err != nil {
		return nil, err
	}
	if err := assertCourierBranch(shipment, eb); err != nil {
		return nil, err
	}
	desc := fmt.Sprintf("Shipment picked up by user with ID %d", userID)
	err = s.repo.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.UpdateShipment(tx, shipment.ID, map[string]any{"delivery_status": StatusPickedUp}); err != nil {
			return err
		}
		if err := s.repo.UpdateShipmentDetail(tx, shipment.ID, map[string]any{
			"pickup_proof": "uploads/photos/" + photoFilename,
		}); err != nil {
			return err
		}
		return s.repo.CreateHistory(tx, &models.ShipmentHistory{
			ShipmentID:  shipment.ID,
			UserID:      &userID,
			BranchID:    &eb.BranchID,
			Status:      StatusPickedUp,
			Description: &desc,
		})
	})
	if err != nil {
		return nil, err
	}
	picked := StatusPickedUp
	shipment.DeliveryStatus = &picked
	return shipment, nil
}

func (s *service) DeliverToBranch(trackingNumber string, userID uint) (*models.Shipment, error) {
	desc := fmt.Sprintf("Shipment picked up by user with ID %d", userID)
	return s.updateCourierStatus(trackingNumber, userID, StatusInTransit, desc)
}

func (s *service) PickShipmentFromBranch(trackingNumber string, userID uint) (*models.Shipment, error) {
	desc := fmt.Sprintf("Shipment picked up by user with ID %d", userID)
	return s.updateCourierStatus(trackingNumber, userID, StatusReadyToDeliver, desc)
}

func (s *service) PickupShipmentFromBranch(trackingNumber string, userID uint) (*models.Shipment, error) {
	desc := fmt.Sprintf("Shipment picked up by user with ID %d", userID)
	return s.updateCourierStatus(trackingNumber, userID, StatusOnTheWayToAddress, desc)
}

func (s *service) DeliverToCustomer(trackingNumber string, userID uint, photoFilename string) (*models.Shipment, error) {
	shipment, eb, err := s.courierShipmentAndBranch(trackingNumber, userID)
	if err != nil {
		return nil, err
	}
	if err := assertCourierBranch(shipment, eb); err != nil {
		return nil, err
	}
	desc := fmt.Sprintf("Shipment picked up by user with ID %d", userID)
	err = s.repo.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.UpdateShipment(tx, shipment.ID, map[string]any{"delivery_status": StatusDelivered}); err != nil {
			return err
		}
		if err := s.repo.UpdateShipmentDetail(tx, shipment.ID, map[string]any{
			"receipt_proof": "uploads/photos/" + photoFilename,
		}); err != nil {
			return err
		}
		return s.repo.CreateHistory(tx, &models.ShipmentHistory{
			ShipmentID:  shipment.ID,
			UserID:      &userID,
			BranchID:    &eb.BranchID,
			Status:      StatusDelivered,
			Description: &desc,
		})
	})
	if err != nil {
		return nil, err
	}
	delivered := StatusDelivered
	shipment.DeliveryStatus = &delivered
	return shipment, nil
}

func (s *service) GeneratePDFByID(id, userID, roleID uint) ([]byte, error) {
	shipment, err := s.FindByID(id, userID, roleID)
	if err != nil {
		return nil, err
	}
	if s.pdfSvc == nil {
		return nil, errors.New("PDF service not available")
	}

	det := shipment.ShipmentDetail
	trackingNumber := ""
	if shipment.TrackingNumber != nil {
		trackingNumber = *shipment.TrackingNumber
	}
	deliveryStatus := ""
	if shipment.DeliveryStatus != nil {
		deliveryStatus = *shipment.DeliveryStatus
	}

	weight := "0"
	basePrice := "0"
	weightPrice := "0"
	distancePrice := "0"
	if det != nil {
		if det.Weight != nil {
			weight = fmt.Sprintf("%.0f", *det.Weight)
		}
		if det.BasePrice != nil {
			basePrice = fmt.Sprintf("%.0f", *det.BasePrice)
		}
		if det.WeightPrice != nil {
			weightPrice = fmt.Sprintf("%.0f", *det.WeightPrice)
		}
		if det.DistancePrice != nil {
			distancePrice = fmt.Sprintf("%.0f", *det.DistancePrice)
		}
	}

	price := "0"
	if shipment.Price != nil {
		price = fmt.Sprintf("%.0f", *shipment.Price)
	}
	distance := "0"
	if shipment.Distance != nil {
		distance = fmt.Sprintf("%.2f", *shipment.Distance)
	}

	qrBase64 := ""
	if shipment.QrCodeImage != nil {
		qrBase64 = s.pdfSvc.QRBase64(*shipment.QrCodeImage)
	}

	senderName, senderEmail, senderPhone, pickupAddress := "", "", "", ""
	recipientName, recipientPhone, destAddress := "", "", ""
	deliveryType, packageType := "", ""
	if det != nil {
		senderName = det.User.Name
		senderEmail = det.User.Email
		senderPhone = det.User.PhoneNumber
		pickupAddress = det.Address.Address
		recipientName = det.RecipientName
		recipientPhone = det.RecipientPhone
		destAddress = det.DestinationAddress
		deliveryType = det.DeliveryType
		packageType = det.PackageType
	}

	data := pdf.ShipmentData{
		TrackingNumber:     trackingNumber,
		ShipmentID:         shipment.ID,
		CreatedDate:        shipment.CreatedAt.Format("02/01/2006"),
		DeliveryType:       deliveryType,
		PackageType:        packageType,
		Weight:             weight,
		Price:              price,
		Distance:           distance,
		PaymentStatus:      shipment.PaymentStatus,
		DeliveryStatus:     deliveryStatus,
		BasePrice:          basePrice,
		WeightPrice:        weightPrice,
		DistancePrice:      distancePrice,
		SenderName:         senderName,
		SenderEmail:        senderEmail,
		SenderPhone:        senderPhone,
		PickupAddress:      pickupAddress,
		RecipientName:      recipientName,
		RecipientPhone:     recipientPhone,
		DestinationAddress: destAddress,
		QRCodeBase64:       qrBase64,
		GeneratedDate:      time.Now().Format("02/01/2006"),
	}
	return s.pdfSvc.Generate(data)
}
