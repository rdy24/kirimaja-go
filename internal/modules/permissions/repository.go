package permissions

import (
	"context"

	"kirimaja-go/models"

	"gorm.io/gorm"
)

type Repository interface {
	FindAll(ctx context.Context) ([]models.Permission, error)
	Create(ctx context.Context, p *models.Permission) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db}
}

func (r *repository) FindAll(ctx context.Context) ([]models.Permission, error) {
	var perms []models.Permission
	err := r.db.WithContext(ctx).Find(&perms).Error
	return perms, err
}

func (r *repository) Create(ctx context.Context, p *models.Permission) error {
	return r.db.WithContext(ctx).Create(p).Error
}
