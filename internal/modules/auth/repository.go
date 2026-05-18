package auth

import (
	"kirimaja-go/models"

	"gorm.io/gorm"
)

type Repository interface {
	FindUserByEmail(email string) (*models.User, error)
	FindRoleByKey(key string) (*models.Role, error)
	CreateUser(user *models.User) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db}
}

func (r *repository) FindUserByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.
		Preload("Role.RolePermissions.Permission").
		Where("email = ?", email).
		First(&user).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &user, err
}

func (r *repository) FindRoleByKey(key string) (*models.Role, error) {
	var role models.Role
	err := r.db.Where("key = ?", key).First(&role).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &role, err
}

func (r *repository) CreateUser(user *models.User) error {
	return r.db.Create(user).Error
}
