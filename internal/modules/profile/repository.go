package profile

import (
	"context"

	"gorm.io/gorm"

	"kirimaja-go/models"
)

type Repository interface {
	FindByID(ctx context.Context, id uint) (*models.User, error)
	Update(ctx context.Context, id uint, data map[string]any) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db}
}

func (r *repository) FindByID(ctx context.Context, id uint) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).First(&user, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &user, err
}

func (r *repository) Update(ctx context.Context, id uint, data map[string]any) error {
	return r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", id).Updates(data).Error
}
