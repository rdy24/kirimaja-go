package roles

import (
	"context"

	"gorm.io/gorm"

	"kirimaja-go/models"
)

type Repository interface {
	FindAll(ctx context.Context) ([]models.Role, error)
	FindByID(ctx context.Context, id uint) (*models.Role, error)
	Create(ctx context.Context, role *models.Role) error
	Update(ctx context.Context, id uint, data map[string]any) error
	Delete(ctx context.Context, id uint) error
	DeletePermissions(ctx context.Context, roleID uint) error
	CreatePermissions(ctx context.Context, roleID uint, permissionIDs []uint) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db}
}

func (r *repository) FindAll(ctx context.Context) ([]models.Role, error) {
	var roles []models.Role
	err := r.db.WithContext(ctx).Preload("RolePermissions.Permission").Find(&roles).Error
	return roles, err
}

func (r *repository) FindByID(ctx context.Context, id uint) (*models.Role, error) {
	var role models.Role
	err := r.db.WithContext(ctx).Preload("RolePermissions.Permission").First(&role, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &role, err
}

func (r *repository) Create(ctx context.Context, role *models.Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}

func (r *repository) Update(ctx context.Context, id uint, data map[string]any) error {
	return r.db.WithContext(ctx).Model(&models.Role{}).Where("id = ?", id).Updates(data).Error
}

func (r *repository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.Role{}, id).Error
}

func (r *repository) DeletePermissions(ctx context.Context, roleID uint) error {
	return r.db.WithContext(ctx).Where("role_id = ?", roleID).Delete(&models.RolePermission{}).Error
}

func (r *repository) CreatePermissions(ctx context.Context, roleID uint, permissionIDs []uint) error {
	if len(permissionIDs) == 0 {
		return nil
	}
	records := make([]models.RolePermission, 0, len(permissionIDs))
	for _, pid := range permissionIDs {
		records = append(records, models.RolePermission{RoleID: roleID, PermissionID: pid})
	}
	return r.db.WithContext(ctx).Create(&records).Error
}
