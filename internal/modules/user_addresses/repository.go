package user_addresses

import (
	"context"

	"kirimaja-go/models"

	"gorm.io/gorm"
)

type Repository interface {
	FindAllByUserID(ctx context.Context, userID uint) ([]models.UserAddress, error)
	FindByID(ctx context.Context, id uint) (*models.UserAddress, error)
	Create(ctx context.Context, addr *models.UserAddress) error
	Update(ctx context.Context, id uint, data map[string]any) error
	Delete(ctx context.Context, id uint) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db}
}

func (r *repository) FindAllByUserID(ctx context.Context, userID uint) ([]models.UserAddress, error) {
	var list []models.UserAddress
	err := r.db.WithContext(ctx).Preload("User").Where("user_id = ?", userID).Find(&list).Error
	return list, err
}

func (r *repository) FindByID(ctx context.Context, id uint) (*models.UserAddress, error) {
	var addr models.UserAddress
	err := r.db.WithContext(ctx).Preload("User").First(&addr, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &addr, err
}

func (r *repository) Create(ctx context.Context, addr *models.UserAddress) error {
	return r.db.WithContext(ctx).Create(addr).Error
}

func (r *repository) Update(ctx context.Context, id uint, data map[string]any) error {
	return r.db.WithContext(ctx).Model(&models.UserAddress{}).Where("id = ?", id).Updates(data).Error
}

func (r *repository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.UserAddress{}, id).Error
}
