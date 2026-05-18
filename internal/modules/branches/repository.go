package branches

import (
	"kirimaja-go/models"

	"gorm.io/gorm"
)

type Repository interface {
	FindAll() ([]models.Branch, error)
	FindByID(id uint) (*models.Branch, error)
	Create(b *models.Branch) error
	Update(id uint, data map[string]any) error
	Delete(id uint) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db}
}

func (r *repository) FindAll() ([]models.Branch, error) {
	var branches []models.Branch
	err := r.db.Find(&branches).Error
	return branches, err
}

func (r *repository) FindByID(id uint) (*models.Branch, error) {
	var branch models.Branch
	err := r.db.First(&branch, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &branch, err
}

func (r *repository) Create(b *models.Branch) error {
	return r.db.Create(b).Error
}

func (r *repository) Update(id uint, data map[string]any) error {
	return r.db.Model(&models.Branch{}).Where("id = ?", id).Updates(data).Error
}

func (r *repository) Delete(id uint) error {
	return r.db.Delete(&models.Branch{}, id).Error
}
