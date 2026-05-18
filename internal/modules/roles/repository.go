package roles

import (
	"kirimaja-go/models"

	"gorm.io/gorm"
)

type Repository interface {
	FindAll() ([]models.Role, error)
	FindByID(id uint) (*models.Role, error)
	Create(role *models.Role) error
	Update(id uint, data map[string]any) error
	Delete(id uint) error
	DeletePermissions(roleID uint) error
	CreatePermissions(roleID uint, permissionIDs []uint) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db}
}

func (r *repository) FindAll() ([]models.Role, error) {
	var roles []models.Role
	err := r.db.Preload("RolePermissions.Permission").Find(&roles).Error
	return roles, err
}

func (r *repository) FindByID(id uint) (*models.Role, error) {
	var role models.Role
	err := r.db.Preload("RolePermissions.Permission").First(&role, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &role, err
}

func (r *repository) Create(role *models.Role) error {
	return r.db.Create(role).Error
}

func (r *repository) Update(id uint, data map[string]any) error {
	return r.db.Model(&models.Role{}).Where("id = ?", id).Updates(data).Error
}

func (r *repository) Delete(id uint) error {
	return r.db.Delete(&models.Role{}, id).Error
}

func (r *repository) DeletePermissions(roleID uint) error {
	return r.db.Where("role_id = ?", roleID).Delete(&models.RolePermission{}).Error
}

func (r *repository) CreatePermissions(roleID uint, permissionIDs []uint) error {
	if len(permissionIDs) == 0 {
		return nil
	}
	records := make([]models.RolePermission, 0, len(permissionIDs))
	for _, pid := range permissionIDs {
		records = append(records, models.RolePermission{RoleID: roleID, PermissionID: pid})
	}
	return r.db.Create(&records).Error
}
