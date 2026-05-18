package history

import (
	"kirimaja-go/models"

	"gorm.io/gorm"
)

type Repository interface {
	FindAll(userID uint, isSuperAdmin bool) ([]models.Shipment, error)
	FindByID(id uint) (*models.Shipment, error)
}

type repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &repository{db} }

func (r *repository) FindAll(userID uint, isSuperAdmin bool) ([]models.Shipment, error) {
	var list []models.Shipment
	q := r.db.
		Preload("ShipmentDetail.User").
		Preload("ShipmentDetail.Address").
		Preload("ShipmentHistories", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC")
		}).
		Where("payment_status = ?", "PAID").
		Order("shipments.created_at DESC")

	if !isSuperAdmin {
		q = q.Where("id IN (SELECT DISTINCT shipment_id FROM shipment_histories WHERE user_id = ?)", userID)
	}
	return list, q.Find(&list).Error
}

func (r *repository) FindByID(id uint) (*models.Shipment, error) {
	var s models.Shipment
	err := r.db.
		Preload("ShipmentDetail.User").
		Preload("ShipmentDetail.Address").
		Preload("ShipmentHistories", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC")
		}).
		Preload("Payment").
		First(&s, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &s, err
}
