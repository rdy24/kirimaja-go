package user_addresses

import (
	"kirimaja-go/models"

	"gorm.io/gorm"
)

type Repository interface {
	FindAllByUserID(userID uint) ([]models.UserAddress, error)
	FindByID(id uint) (*models.UserAddress, error)
	Create(addr *models.UserAddress) error
	Update(id uint, data map[string]any) error
	Delete(id uint) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db}
}

func (r *repository) FindAllByUserID(userID uint) ([]models.UserAddress, error) {
	var list []models.UserAddress
	err := r.db.Preload("User").Where("user_id = ?", userID).Find(&list).Error
	return list, err
}

func (r *repository) FindByID(id uint) (*models.UserAddress, error) {
	var addr models.UserAddress
	err := r.db.Preload("User").First(&addr, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &addr, err
}

func (r *repository) Create(addr *models.UserAddress) error {
	return r.db.Create(addr).Error
}

func (r *repository) Update(id uint, data map[string]any) error {
	return r.db.Model(&models.UserAddress{}).Where("id = ?", id).Updates(data).Error
}

func (r *repository) Delete(id uint) error {
	return r.db.Delete(&models.UserAddress{}, id).Error
}
