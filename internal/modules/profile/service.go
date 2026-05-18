package profile

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	FindOne(userID uint) (*ProfileResponse, error)
	Update(userID uint, req UpdateProfileRequest) (*ProfileResponse, error)
	UpdateAvatar(userID uint, avatarPath string) (*ProfileResponse, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo}
}

func (s *service) FindOne(userID uint) (*ProfileResponse, error) {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("user with id %d not found", userID)
	}
	return &ProfileResponse{
		ID:          user.ID,
		Name:        user.Name,
		Email:       user.Email,
		Avatar:      user.Avatar,
		PhoneNumber: user.PhoneNumber,
	}, nil
}

func (s *service) Update(userID uint, req UpdateProfileRequest) (*ProfileResponse, error) {
	user, err := s.repo.FindByID(userID)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}

	data := map[string]any{}
	if req.Name != "" {
		data["name"] = req.Name
	}
	if req.Email != "" {
		data["email"] = req.Email
	}
	if req.PhoneNumber != "" {
		data["phone_number"] = req.PhoneNumber
	}
	if req.Password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), 10)
		if err != nil {
			return nil, errors.New("failed to hash password")
		}
		data["password"] = string(hashed)
	}

	if len(data) > 0 {
		if err := s.repo.Update(userID, data); err != nil {
			return nil, err
		}
	}
	return s.FindOne(userID)
}

func (s *service) UpdateAvatar(userID uint, avatarPath string) (*ProfileResponse, error) {
	user, err := s.repo.FindByID(userID)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}
	if err := s.repo.Update(userID, map[string]any{"avatar": avatarPath}); err != nil {
		return nil, err
	}
	return s.FindOne(userID)
}
