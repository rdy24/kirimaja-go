package roles

import (
	"errors"
	"fmt"

	"kirimaja-go/models"
)

type Service interface {
	FindAll() ([]RoleResponse, error)
	FindByID(id uint) (*RoleResponse, error)
	Create(req CreateRoleRequest) (*RoleResponse, error)
	Update(id uint, req UpdateRoleRequest) (*RoleResponse, error)
	Delete(id uint) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo}
}

func (s *service) FindAll() ([]RoleResponse, error) {
	roles, err := s.repo.FindAll()
	if err != nil {
		return nil, err
	}
	result := make([]RoleResponse, 0, len(roles))
	for _, r := range roles {
		result = append(result, toRoleResponse(r))
	}
	return result, nil
}

func (s *service) FindByID(id uint) (*RoleResponse, error) {
	role, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, fmt.Errorf("role with ID %d not found", id)
	}
	res := toRoleResponse(*role)
	return &res, nil
}

func (s *service) Create(req CreateRoleRequest) (*RoleResponse, error) {
	role := &models.Role{Name: req.Name, Key: req.Key}
	if err := s.repo.Create(role); err != nil {
		return nil, err
	}
	return s.FindByID(role.ID)
}

func (s *service) Update(id uint, req UpdateRoleRequest) (*RoleResponse, error) {
	if _, err := s.FindByID(id); err != nil {
		return nil, errors.New("role tidak ditemukan")
	}
	if err := s.repo.DeletePermissions(id); err != nil {
		return nil, err
	}
	if err := s.repo.CreatePermissions(id, req.PermissionIDs); err != nil {
		return nil, err
	}
	return s.FindByID(id)
}

func (s *service) Delete(id uint) error {
	if _, err := s.FindByID(id); err != nil {
		return errors.New("role tidak ditemukan")
	}
	return s.repo.Delete(id)
}

func toRoleResponse(r models.Role) RoleResponse {
	perms := make([]PermissionResponse, 0, len(r.RolePermissions))
	for _, rp := range r.RolePermissions {
		perms = append(perms, PermissionResponse{
			ID:       rp.Permission.ID,
			Name:     rp.Permission.Name,
			Key:      rp.Permission.Key,
			Resource: rp.Permission.Resource,
		})
	}
	return RoleResponse{ID: r.ID, Name: r.Name, Key: r.Key, Permissions: perms}
}
