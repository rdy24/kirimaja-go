package profile

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	FindOne(ctx context.Context, userID uint) (*ProfileResponse, error)
	Update(ctx context.Context, userID uint, req UpdateProfileRequest) (*ProfileResponse, error)
	UpdateAvatar(ctx context.Context, userID uint, avatarPath string) (*ProfileResponse, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo}
}

func (s *service) FindOne(ctx context.Context, userID uint) (*ProfileResponse, error) {
	user, err := s.repo.FindByID(ctx, userID)
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

func (s *service) Update(ctx context.Context, userID uint, req UpdateProfileRequest) (*ProfileResponse, error) {
	user, err := s.repo.FindByID(ctx, userID)
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
		if err := s.repo.Update(ctx, userID, data); err != nil {
			return nil, err
		}
	}
	return s.FindOne(ctx, userID)
}

func (s *service) UpdateAvatar(ctx context.Context, userID uint, avatarPath string) (*ProfileResponse, error) {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}
	if err := s.repo.Update(ctx, userID, map[string]any{"avatar": avatarPath}); err != nil {
		return nil, err
	}
	return s.FindOne(ctx, userID)
}
