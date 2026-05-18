package branches

import (
	"context"
	"errors"
	"fmt"

	"kirimaja-go/models"
)

type Service interface {
	FindAll(ctx context.Context) ([]BranchResponse, error)
	FindByID(ctx context.Context, id uint) (*BranchResponse, error)
	Create(ctx context.Context, req CreateBranchRequest) (*BranchResponse, error)
	Update(ctx context.Context, id uint, req UpdateBranchRequest) (*BranchResponse, error)
	Delete(ctx context.Context, id uint) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo}
}

func (s *service) FindAll(ctx context.Context) ([]BranchResponse, error) {
	list, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]BranchResponse, 0, len(list))
	for _, b := range list {
		result = append(result, toBranchResponse(b))
	}
	return result, nil
}

func (s *service) FindByID(ctx context.Context, id uint) (*BranchResponse, error) {
	b, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if b == nil {
		return nil, fmt.Errorf("branch with ID %d not found", id)
	}
	res := toBranchResponse(*b)
	return &res, nil
}

func (s *service) Create(ctx context.Context, req CreateBranchRequest) (*BranchResponse, error) {
	b := &models.Branch{Name: req.Name, Address: req.Address, PhoneNumber: req.PhoneNumber}
	if err := s.repo.Create(ctx, b); err != nil {
		return nil, err
	}
	res := toBranchResponse(*b)
	return &res, nil
}

func (s *service) Update(ctx context.Context, id uint, req UpdateBranchRequest) (*BranchResponse, error) {
	if _, err := s.FindByID(ctx, id); err != nil {
		return nil, errors.New("branch tidak ditemukan")
	}
	data := map[string]any{}
	if req.Name != "" {
		data["name"] = req.Name
	}
	if req.Address != "" {
		data["address"] = req.Address
	}
	if req.PhoneNumber != "" {
		data["phone_number"] = req.PhoneNumber
	}
	if err := s.repo.Update(ctx, id, data); err != nil {
		return nil, err
	}
	return s.FindByID(ctx, id)
}

func (s *service) Delete(ctx context.Context, id uint) error {
	if _, err := s.FindByID(ctx, id); err != nil {
		return errors.New("branch tidak ditemukan")
	}
	return s.repo.Delete(ctx, id)
}

func toBranchResponse(b models.Branch) BranchResponse {
	return BranchResponse{ID: b.ID, Name: b.Name, Address: b.Address, PhoneNumber: b.PhoneNumber}
}
