package permissions

import (
	"context"

	"kirimaja-go/models"
)

type Service interface {
	FindAll(ctx context.Context) ([]PermissionResponse, error)
	Create(ctx context.Context, req CreatePermissionRequest) (*PermissionResponse, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo}
}

func (s *service) FindAll(ctx context.Context) ([]PermissionResponse, error) {
	perms, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]PermissionResponse, 0, len(perms))
	for _, p := range perms {
		result = append(result, PermissionResponse{ID: p.ID, Name: p.Name, Key: p.Key, Resource: p.Resource})
	}
	return result, nil
}

func (s *service) Create(ctx context.Context, req CreatePermissionRequest) (*PermissionResponse, error) {
	p := &models.Permission{Name: req.Name, Key: req.Key, Resource: req.Resource}
	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}
	return &PermissionResponse{ID: p.ID, Name: p.Name, Key: p.Key, Resource: p.Resource}, nil
}
