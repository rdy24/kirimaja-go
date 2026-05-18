package shipments

import (
	"context"

	"gorm.io/gorm"

	"kirimaja-go/models"
)

// Repository keeps GORM entirely behind this interface. Multi-step writes go
// through Atomic, which runs the callback inside a single transaction with a
// transactional Repository — the service layer never sees *gorm.DB.
//
// Every method takes context.Context so a cancelled/timed-out request (client
// disconnect, deadline) actually cancels the DB work instead of running on.
type Repository interface {
	// Atomic runs fn inside one DB transaction. The Repository passed to fn
	// is transaction-scoped; commit on nil error, rollback otherwise.
	Atomic(ctx context.Context, fn func(Repository) error) error

	FindAll(ctx context.Context, userID uint) ([]models.Shipment, error)
	FindByID(ctx context.Context, id uint) (*models.Shipment, error)
	FindByTrackingNumber(ctx context.Context, tracking string) (*models.Shipment, error)
	FindPickupAddress(ctx context.Context, id uint) (*models.UserAddress, error)
	CreateShipment(ctx context.Context, shipment *models.Shipment) error
	CreateShipmentDetail(ctx context.Context, detail *models.ShipmentDetail) error
	CreatePayment(ctx context.Context, payment *models.Payment) error
	CreateHistory(ctx context.Context, history *models.ShipmentHistory) error
	UpdateShipment(ctx context.Context, id uint, data map[string]any) error
	UpdatePayment(ctx context.Context, id uint, data map[string]any) error
	UpdateShipmentDetail(ctx context.Context, shipmentID uint, data map[string]any) error
	FindPaymentByExternalID(ctx context.Context, externalID string) (*models.Payment, error)

	// Branch / courier
	FindEmployeeBranch(ctx context.Context, userID uint) (*models.EmployeeBranch, error)
	FindAllBranchLogs(ctx context.Context, branchID *uint) ([]models.ShipmentBranchLog, error)
	CreateBranchLog(ctx context.Context, log *models.ShipmentBranchLog) (*models.ShipmentBranchLog, error)
	FindLastBranchLogIn(ctx context.Context, trackingNumber string, branchID uint) (*models.ShipmentBranchLog, error)
	FindAllForCourier(ctx context.Context) ([]models.Shipment, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db}
}

func (r *repository) Atomic(ctx context.Context, fn func(Repository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(&repository{db: tx})
	})
}

func (r *repository) FindAll(ctx context.Context, userID uint) ([]models.Shipment, error) {
	var list []models.Shipment
	err := r.db.WithContext(ctx).
		Joins("JOIN shipment_details sd ON sd.shipment_id = shipments.id AND sd.user_id = ?", userID).
		Preload("ShipmentDetail").
		Preload("Payment").
		Preload("ShipmentHistories", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC")
		}).
		Order("shipments.created_at DESC").
		Find(&list).Error
	return list, err
}

func (r *repository) FindByID(ctx context.Context, id uint) (*models.Shipment, error) {
	var s models.Shipment
	err := r.db.WithContext(ctx).
		Preload("ShipmentDetail.User").
		Preload("ShipmentDetail.Address").
		Preload("Payment").
		Preload("ShipmentHistories", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC")
		}).
		First(&s, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &s, err
}

func (r *repository) FindByTrackingNumber(ctx context.Context, tracking string) (*models.Shipment, error) {
	var s models.Shipment
	err := r.db.WithContext(ctx).
		Preload("ShipmentDetail.User").
		Preload("ShipmentDetail.Address").
		Preload("Payment").
		Preload("ShipmentHistories", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC")
		}).
		Where("tracking_number = ?", tracking).
		First(&s).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &s, err
}

func (r *repository) FindPickupAddress(ctx context.Context, id uint) (*models.UserAddress, error) {
	var addr models.UserAddress
	err := r.db.WithContext(ctx).Preload("User").First(&addr, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &addr, err
}

func (r *repository) CreateShipment(ctx context.Context, shipment *models.Shipment) error {
	return r.db.WithContext(ctx).Create(shipment).Error
}

func (r *repository) CreateShipmentDetail(ctx context.Context, detail *models.ShipmentDetail) error {
	return r.db.WithContext(ctx).Create(detail).Error
}

func (r *repository) CreatePayment(ctx context.Context, payment *models.Payment) error {
	return r.db.WithContext(ctx).Create(payment).Error
}

func (r *repository) CreateHistory(ctx context.Context, history *models.ShipmentHistory) error {
	return r.db.WithContext(ctx).Create(history).Error
}

func (r *repository) UpdateShipment(ctx context.Context, id uint, data map[string]any) error {
	return r.db.WithContext(ctx).Model(&models.Shipment{}).Where("id = ?", id).Updates(data).Error
}

func (r *repository) UpdatePayment(ctx context.Context, id uint, data map[string]any) error {
	return r.db.WithContext(ctx).Model(&models.Payment{}).Where("id = ?", id).Updates(data).Error
}

func (r *repository) UpdateShipmentDetail(ctx context.Context, shipmentID uint, data map[string]any) error {
	return r.db.WithContext(ctx).Model(&models.ShipmentDetail{}).Where("shipment_id = ?", shipmentID).Updates(data).Error
}

func (r *repository) FindPaymentByExternalID(ctx context.Context, externalID string) (*models.Payment, error) {
	var p models.Payment
	err := r.db.WithContext(ctx).
		Preload("Shipment.ShipmentDetail.User").
		Where("external_id = ?", externalID).
		First(&p).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &p, err
}

func (r *repository) FindEmployeeBranch(ctx context.Context, userID uint) (*models.EmployeeBranch, error) {
	var eb models.EmployeeBranch
	err := r.db.WithContext(ctx).Preload("Branch").Where("user_id = ?", userID).First(&eb).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &eb, err
}

func (r *repository) FindAllBranchLogs(ctx context.Context, branchID *uint) ([]models.ShipmentBranchLog, error) {
	var logs []models.ShipmentBranchLog
	q := r.db.WithContext(ctx).Preload("Shipment.ShipmentDetail").Preload("Branch").Preload("ScannedByUser").
		Order("created_at DESC")
	if branchID != nil {
		q = q.Where("branch_id = ?", *branchID)
	}
	return logs, q.Find(&logs).Error
}

func (r *repository) CreateBranchLog(ctx context.Context, log *models.ShipmentBranchLog) (*models.ShipmentBranchLog, error) {
	if err := r.db.WithContext(ctx).Create(log).Error; err != nil {
		return nil, err
	}
	if err := r.db.WithContext(ctx).Preload("Shipment.ShipmentDetail").Preload("Branch").Preload("ScannedByUser").
		First(log, log.ID).Error; err != nil {
		return nil, err
	}
	return log, nil
}

func (r *repository) FindLastBranchLogIn(ctx context.Context, trackingNumber string, branchID uint) (*models.ShipmentBranchLog, error) {
	var log models.ShipmentBranchLog
	err := r.db.WithContext(ctx).
		Where("tracking_number = ? AND type = 'IN' AND branch_id = ?", trackingNumber, branchID).
		Order("created_at DESC").
		First(&log).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &log, err
}

func (r *repository) FindAllForCourier(ctx context.Context) ([]models.Shipment, error) {
	var list []models.Shipment
	statuses := []string{
		StatusReadyToPickup, StatusWaitingPickup, StatusPickedUp,
		StatusReadyToPickupAtBranch, StatusReadyToDeliver,
		StatusOnTheWayToAddress, StatusOnTheWay, StatusDelivered,
	}
	err := r.db.WithContext(ctx).
		Where("payment_status = ? AND delivery_status IN ?", StatusPaid, statuses).
		Preload("ShipmentDetail.User").
		Preload("ShipmentDetail.Address").
		Preload("ShipmentHistories").
		Preload("Payment").
		Order("created_at DESC").
		Find(&list).Error
	return list, err
}
