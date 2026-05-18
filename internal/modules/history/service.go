package history

import (
	"context"
	"errors"

	"kirimaja-go/models"
)

const superAdminRoleID uint = 1

type Service interface {
	FindAll(ctx context.Context, userID, roleID uint) ([]models.Shipment, error)
	FindByID(ctx context.Context, id uint) (*models.Shipment, error)
}

type service struct{ repo Repository }

func NewService(repo Repository) Service { return &service{repo} }

func (s *service) FindAll(ctx context.Context, userID, roleID uint) ([]models.Shipment, error) {
	return s.repo.FindAll(ctx, userID, roleID == superAdminRoleID)
}

func (s *service) FindByID(ctx context.Context, id uint) (*models.Shipment, error) {
	shipment, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if shipment == nil {
		return nil, errors.New("shipment not found")
	}
	return shipment, nil
}
