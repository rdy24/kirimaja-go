package profile

import (
	"kirimaja-go/models"

	"gorm.io/gorm"
)

type Repository interface {
	FindByID(id uint) (*models.User, error)
	Update(id uint, data map[string]any) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db}
}

func (r *repository) FindByID(id uint) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &user, err
}

func (r *repository) Update(id uint, data map[string]any) error {
	return r.db.Model(&models.User{}).Where("id = ?", id).Updates(data).Error
}
