package roles

import (
	"context"
	"errors"
	"fmt"

	"kirimaja-go/models"
)

type Service interface {
	FindAll(ctx context.Context) ([]RoleResponse, error)
	FindByID(ctx context.Context, id uint) (*RoleResponse, error)
	Create(ctx context.Context, req CreateRoleRequest) (*RoleResponse, error)
	Update(ctx context.Context, id uint, req UpdateRoleRequest) (*RoleResponse, error)
	Delete(ctx context.Context, id uint) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo}
}

func (s *service) FindAll(ctx context.Context) ([]RoleResponse, error) {
	roles, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]RoleResponse, 0, len(roles))
	for _, r := range roles {
		result = append(result, toRoleResponse(r))
	}
	return result, nil
}

func (s *service) FindByID(ctx context.Context, id uint) (*RoleResponse, error) {
	role, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, fmt.Errorf("role with ID %d not found", id)
	}
	res := toRoleResponse(*role)
	return &res, nil
}

func (s *service) Create(ctx context.Context, req CreateRoleRequest) (*RoleResponse, error) {
	role := &models.Role{Name: req.Name, Key: req.Key}
	if err := s.repo.Create(ctx, role); err != nil {
		return nil, err
	}
	return s.FindByID(ctx, role.ID)
}

func (s *service) Update(ctx context.Context, id uint, req UpdateRoleRequest) (*RoleResponse, error) {
	if _, err := s.FindByID(ctx, id); err != nil {
		return nil, errors.New("role tidak ditemukan")
	}
	if err := s.repo.DeletePermissions(ctx, id); err != nil {
		return nil, err
	}
	if err := s.repo.CreatePermissions(ctx, id, req.PermissionIDs); err != nil {
		return nil, err
	}
	return s.FindByID(ctx, id)
}

func (s *service) Delete(ctx context.Context, id uint) error {
	if _, err := s.FindByID(ctx, id); err != nil {
		return errors.New("role tidak ditemukan")
	}
	return s.repo.Delete(ctx, id)
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
