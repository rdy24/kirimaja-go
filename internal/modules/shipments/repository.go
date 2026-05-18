package shipments

import (
	"kirimaja-go/models"

	"gorm.io/gorm"
)

type Repository interface {
	FindAll(userID uint) ([]models.Shipment, error)
	FindByID(id uint) (*models.Shipment, error)
	FindByTrackingNumber(tracking string) (*models.Shipment, error)
	FindPickupAddress(id uint) (*models.UserAddress, error)
	Transaction(fn func(tx *gorm.DB) error) error
	CreateShipment(tx *gorm.DB, shipment *models.Shipment) error
	CreateShipmentDetail(tx *gorm.DB, detail *models.ShipmentDetail) error
	CreatePayment(tx *gorm.DB, payment *models.Payment) error
	CreateHistory(tx *gorm.DB, history *models.ShipmentHistory) error
	UpdateShipment(tx *gorm.DB, id uint, data map[string]any) error
	UpdatePayment(tx *gorm.DB, id uint, data map[string]any) error
	UpdateShipmentDetail(tx *gorm.DB, shipmentID uint, data map[string]any) error
	FindPaymentByExternalID(externalID string) (*models.Payment, error)

	// Branch / courier
	FindEmployeeBranch(userID uint) (*models.EmployeeBranch, error)
	FindAllBranchLogs(branchID *uint) ([]models.ShipmentBranchLog, error)
	CreateBranchLog(tx *gorm.DB, log *models.ShipmentBranchLog) (*models.ShipmentBranchLog, error)
	FindLastBranchLogIn(trackingNumber string, branchID uint) (*models.ShipmentBranchLog, error)
	FindAllForCourier() ([]models.Shipment, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db}
}

func (r *repository) FindAll(userID uint) ([]models.Shipment, error) {
	var list []models.Shipment
	err := r.db.
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

func (r *repository) FindByID(id uint) (*models.Shipment, error) {
	var s models.Shipment
	err := r.db.
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

func (r *repository) FindByTrackingNumber(tracking string) (*models.Shipment, error) {
	var s models.Shipment
	err := r.db.
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

func (r *repository) FindPickupAddress(id uint) (*models.UserAddress, error) {
	var addr models.UserAddress
	err := r.db.Preload("User").First(&addr, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &addr, err
}

func (r *repository) Transaction(fn func(tx *gorm.DB) error) error {
	return r.db.Transaction(fn)
}

func (r *repository) CreateShipment(tx *gorm.DB, shipment *models.Shipment) error {
	return tx.Create(shipment).Error
}

func (r *repository) CreateShipmentDetail(tx *gorm.DB, detail *models.ShipmentDetail) error {
	return tx.Create(detail).Error
}

func (r *repository) CreatePayment(tx *gorm.DB, payment *models.Payment) error {
	return tx.Create(payment).Error
}

func (r *repository) CreateHistory(tx *gorm.DB, history *models.ShipmentHistory) error {
	return tx.Create(history).Error
}

func (r *repository) UpdateShipment(tx *gorm.DB, id uint, data map[string]any) error {
	return tx.Model(&models.Shipment{}).Where("id = ?", id).Updates(data).Error
}

func (r *repository) UpdatePayment(tx *gorm.DB, id uint, data map[string]any) error {
	return tx.Model(&models.Payment{}).Where("id = ?", id).Updates(data).Error
}

func (r *repository) UpdateShipmentDetail(tx *gorm.DB, shipmentID uint, data map[string]any) error {
	return tx.Model(&models.ShipmentDetail{}).Where("shipment_id = ?", shipmentID).Updates(data).Error
}

func (r *repository) FindPaymentByExternalID(externalID string) (*models.Payment, error) {
	var p models.Payment
	err := r.db.
		Preload("Shipment.ShipmentDetail.User").
		Where("external_id = ?", externalID).
		First(&p).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &p, err
}

func (r *repository) FindEmployeeBranch(userID uint) (*models.EmployeeBranch, error) {
	var eb models.EmployeeBranch
	err := r.db.Preload("Branch").Where("user_id = ?", userID).First(&eb).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &eb, err
}

func (r *repository) FindAllBranchLogs(branchID *uint) ([]models.ShipmentBranchLog, error) {
	var logs []models.ShipmentBranchLog
	q := r.db.Preload("Shipment.ShipmentDetail").Preload("Branch").Preload("ScannedByUser").
		Order("created_at DESC")
	if branchID != nil {
		q = q.Where("branch_id = ?", *branchID)
	}
	return logs, q.Find(&logs).Error
}

func (r *repository) CreateBranchLog(tx *gorm.DB, log *models.ShipmentBranchLog) (*models.ShipmentBranchLog, error) {
	if err := tx.Create(log).Error; err != nil {
		return nil, err
	}
	if err := tx.Preload("Shipment.ShipmentDetail").Preload("Branch").Preload("ScannedByUser").
		First(log, log.ID).Error; err != nil {
		return nil, err
	}
	return log, nil
}

func (r *repository) FindLastBranchLogIn(trackingNumber string, branchID uint) (*models.ShipmentBranchLog, error) {
	var log models.ShipmentBranchLog
	err := r.db.
		Where("tracking_number = ? AND type = 'IN' AND branch_id = ?", trackingNumber, branchID).
		Order("created_at DESC").
		First(&log).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &log, err
}

func (r *repository) FindAllForCourier() ([]models.Shipment, error) {
	var list []models.Shipment
	statuses := []string{
		StatusReadyToPickup, StatusWaitingPickup, StatusPickedUp,
		StatusReadyToPickupAtBranch, StatusReadyToDeliver,
		StatusOnTheWayToAddress, StatusOnTheWay, StatusDelivered,
	}
	err := r.db.
		Where("payment_status = ? AND delivery_status IN ?", StatusPaid, statuses).
		Preload("ShipmentDetail.User").
		Preload("ShipmentDetail.Address").
		Preload("ShipmentHistories").
		Preload("Payment").
		Order("created_at DESC").
		Find(&list).Error
	return list, err
}
