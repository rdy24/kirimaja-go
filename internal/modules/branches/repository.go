package branches

import (
	"context"

	"kirimaja-go/models"

	"gorm.io/gorm"
)

type Repository interface {
	FindAll(ctx context.Context) ([]models.Branch, error)
	FindByID(ctx context.Context, id uint) (*models.Branch, error)
	Create(ctx context.Context, b *models.Branch) error
	Update(ctx context.Context, id uint, data map[string]any) error
	Delete(ctx context.Context, id uint) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db}
}

func (r *repository) FindAll(ctx context.Context) ([]models.Branch, error) {
	var branches []models.Branch
	err := r.db.WithContext(ctx).Find(&branches).Error
	return branches, err
}

func (r *repository) FindByID(ctx context.Context, id uint) (*models.Branch, error) {
	var branch models.Branch
	err := r.db.WithContext(ctx).First(&branch, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &branch, err
}

func (r *repository) Create(ctx context.Context, b *models.Branch) error {
	return r.db.WithContext(ctx).Create(b).Error
}

func (r *repository) Update(ctx context.Context, id uint, data map[string]any) error {
	return r.db.WithContext(ctx).Model(&models.Branch{}).Where("id = ?", id).Updates(data).Error
}

func (r *repository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.Branch{}, id).Error
}
