package auth

import (
	"context"

	"gorm.io/gorm"

	"kirimaja-go/models"
)

type Repository interface {
	FindUserByEmail(ctx context.Context, email string) (*models.User, error)
	FindRoleByKey(ctx context.Context, key string) (*models.Role, error)
	CreateUser(ctx context.Context, user *models.User) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db}
}

func (r *repository) FindUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).
		Preload("Role.RolePermissions.Permission").
		Where("email = ?", email).
		First(&user).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &user, err
}

func (r *repository) FindRoleByKey(ctx context.Context, key string) (*models.Role, error) {
	var role models.Role
	err := r.db.WithContext(ctx).Where("key = ?", key).First(&role).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &role, err
}

func (r *repository) CreateUser(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}
