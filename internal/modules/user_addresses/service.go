package user_addresses

import (
	"errors"
	"fmt"

	"kirimaja-go/internal/common/opencage"
	"kirimaja-go/models"
)

type Service interface {
	FindAll(userID uint) ([]UserAddressResponse, error)
	FindByID(id uint) (*UserAddressResponse, error)
	Create(userID uint, req CreateUserAddressRequest, photoPath *string) (*UserAddressResponse, error)
	Update(id uint, req UpdateUserAddressRequest, photoPath *string) (*UserAddressResponse, error)
	Delete(id uint) error
}

type service struct {
	repo   Repository
	geocli *opencage.Client
}

func NewService(repo Repository, geocli *opencage.Client) Service {
	return &service{repo, geocli}
}

func (s *service) FindAll(userID uint) ([]UserAddressResponse, error) {
	list, err := s.repo.FindAllByUserID(userID)
	if err != nil {
		return nil, err
	}
	result := make([]UserAddressResponse, 0, len(list))
	for _, a := range list {
		result = append(result, toResponse(a))
	}
	return result, nil
}

func (s *service) FindByID(id uint) (*UserAddressResponse, error) {
	addr, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if addr == nil {
		return nil, fmt.Errorf("user address with ID %d not found", id)
	}
	res := toResponse(*addr)
	return &res, nil
}

func (s *service) Create(userID uint, req CreateUserAddressRequest, photoPath *string) (*UserAddressResponse, error) {
	loc, err := s.geocli.Geocode(req.Address)
	if err != nil {
		return nil, fmt.Errorf("geocode gagal: %w", err)
	}

	photo := req.Photo
	if photoPath != nil {
		photo = photoPath
	}

	addr := &models.UserAddress{
		UserID:    userID,
		Address:   req.Address,
		Tag:       &req.Tag,
		Label:     &req.Label,
		Photo:     photo,
		Latitude:  &loc.Lat,
		Longitude: &loc.Lng,
	}
	if err := s.repo.Create(addr); err != nil {
		return nil, err
	}
	return s.FindByID(addr.ID)
}

func (s *service) Update(id uint, req UpdateUserAddressRequest, photoPath *string) (*UserAddressResponse, error) {
	existing, err := s.repo.FindByID(id)
	if err != nil || existing == nil {
		return nil, errors.New("user address tidak ditemukan")
	}

	data := map[string]any{}

	if req.Address != "" && req.Address != existing.Address {
		loc, err := s.geocli.Geocode(req.Address)
		if err != nil {
			return nil, fmt.Errorf("geocode gagal: %w", err)
		}
		data["address"] = req.Address
		data["latitude"] = loc.Lat
		data["longitude"] = loc.Lng
	}
	if req.Tag != "" {
		data["tag"] = req.Tag
	}
	if req.Label != "" {
		data["label"] = req.Label
	}
	if photoPath != nil {
		data["photo"] = *photoPath
	} else if req.Photo != nil {
		data["photo"] = *req.Photo
	}

	if len(data) > 0 {
		if err := s.repo.Update(id, data); err != nil {
			return nil, err
		}
	}
	return s.FindByID(id)
}

func (s *service) Delete(id uint) error {
	if _, err := s.FindByID(id); err != nil {
		return err
	}
	return s.repo.Delete(id)
}

func toResponse(a models.UserAddress) UserAddressResponse {
	return UserAddressResponse{
		ID:        a.ID,
		UserID:    a.UserID,
		Address:   a.Address,
		Tag:       a.Tag,
		Label:     a.Label,
		Photo:     a.Photo,
		Latitude:  a.Latitude,
		Longitude: a.Longitude,
		User: UserInfo{
			ID:          a.User.ID,
			Name:        a.User.Name,
			Email:       a.User.Email,
			PhoneNumber: a.User.PhoneNumber,
			Avatar:      a.User.Avatar,
		},
	}
}
