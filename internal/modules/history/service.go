package history

import (
	"errors"

	"kirimaja-go/models"
)

const superAdminRoleID uint = 1

type Service interface {
	FindAll(userID, roleID uint) ([]models.Shipment, error)
	FindByID(id uint) (*models.Shipment, error)
}

type service struct{ repo Repository }

func NewService(repo Repository) Service { return &service{repo} }

func (s *service) FindAll(userID, roleID uint) ([]models.Shipment, error) {
	return s.repo.FindAll(userID, roleID == superAdminRoleID)
}

func (s *service) FindByID(id uint) (*models.Shipment, error) {
	shipment, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if shipment == nil {
		return nil, errors.New("shipment not found")
	}
	return shipment, nil
}
