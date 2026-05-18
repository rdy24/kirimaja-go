package permissions

import "kirimaja-go/models"

type Service interface {
	FindAll() ([]PermissionResponse, error)
	Create(req CreatePermissionRequest) (*PermissionResponse, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo}
}

func (s *service) FindAll() ([]PermissionResponse, error) {
	perms, err := s.repo.FindAll()
	if err != nil {
		return nil, err
	}
	result := make([]PermissionResponse, 0, len(perms))
	for _, p := range perms {
		result = append(result, PermissionResponse{ID: p.ID, Name: p.Name, Key: p.Key, Resource: p.Resource})
	}
	return result, nil
}

func (s *service) Create(req CreatePermissionRequest) (*PermissionResponse, error) {
	p := &models.Permission{Name: req.Name, Key: req.Key, Resource: req.Resource}
	if err := s.repo.Create(p); err != nil {
		return nil, err
	}
	return &PermissionResponse{ID: p.ID, Name: p.Name, Key: p.Key, Resource: p.Resource}, nil
}
