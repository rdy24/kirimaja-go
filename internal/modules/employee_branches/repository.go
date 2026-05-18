package employee_branches

import (
	"kirimaja-go/models"

	"gorm.io/gorm"
)

type Repository interface {
	FindAll() ([]models.EmployeeBranch, error)
	FindByID(id uint) (*models.EmployeeBranch, error)
	FindUserByEmail(email string) (*models.User, error)
	FindBranchByID(id uint) (*models.Branch, error)
	FindRoleByID(id uint) (*models.Role, error)
	CreateWithUser(user *models.User, eb *models.EmployeeBranch) error
	UpdateWithUser(eb *models.EmployeeBranch, userData map[string]any, ebData map[string]any) error
	DeleteWithUser(eb *models.EmployeeBranch) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db}
}

func (r *repository) FindAll() ([]models.EmployeeBranch, error) {
	var list []models.EmployeeBranch
	err := r.db.Preload("User").Preload("Branch").Find(&list).Error
	return list, err
}

func (r *repository) FindByID(id uint) (*models.EmployeeBranch, error) {
	var eb models.EmployeeBranch
	err := r.db.Preload("User").Preload("Branch").First(&eb, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &eb, err
}

func (r *repository) FindUserByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &user, err
}

func (r *repository) FindBranchByID(id uint) (*models.Branch, error) {
	var branch models.Branch
	err := r.db.First(&branch, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &branch, err
}

func (r *repository) FindRoleByID(id uint) (*models.Role, error) {
	var role models.Role
	err := r.db.First(&role, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &role, err
}

func (r *repository) CreateWithUser(user *models.User, eb *models.EmployeeBranch) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		eb.UserID = user.ID
		return tx.Create(eb).Error
	})
}

func (r *repository) UpdateWithUser(eb *models.EmployeeBranch, userData map[string]any, ebData map[string]any) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if len(userData) > 0 {
			if err := tx.Model(&models.User{}).Where("id = ?", eb.UserID).Updates(userData).Error; err != nil {
				return err
			}
		}
		if len(ebData) > 0 {
			if err := tx.Model(&models.EmployeeBranch{}).Where("id = ?", eb.ID).Updates(ebData).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *repository) DeleteWithUser(eb *models.EmployeeBranch) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&models.EmployeeBranch{}, eb.ID).Error; err != nil {
			return err
		}
		return tx.Delete(&models.User{}, eb.UserID).Error
	})
}
