package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"kirimaja-go/models"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrEmailRegistered    = errors.New("email already registered")
)

type Service interface {
	Login(ctx context.Context, req LoginRequest) (*AuthResponse, error)
	Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error)
}

type service struct {
	repo         Repository
	jwtSecret    string
	jwtExpiresIn string
}

func NewService(repo Repository, jwtSecret, jwtExpiresIn string) Service {
	return &service{repo, jwtSecret, jwtExpiresIn}
}

func (s *service) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	user, err := s.repo.FindUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}

	return buildAuthResponse(token, user), nil
}

func (s *service) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	existing, err := s.repo.FindUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrEmailRegistered
	}

	role, err := s.repo.FindRoleByKey(ctx, "customer")
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, errors.New("role customer tidak ditemukan")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), 10)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &models.User{
		Name:        req.Name,
		Email:       req.Email,
		Password:    string(hashed),
		PhoneNumber: req.PhoneNumber,
		RoleID:      role.ID,
	}
	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	// Reload user with role + permissions
	user, err = s.repo.FindUserByEmail(ctx, req.Email)
	if err != nil || user == nil {
		return nil, errors.New("failed to load created user")
	}

	token, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}

	return buildAuthResponse(token, user), nil
}

func (s *service) generateToken(user *models.User) (string, error) {
	dur, err := time.ParseDuration(s.jwtExpiresIn)
	if err != nil {
		dur = 24 * time.Hour
	}

	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"name":    user.Name,
		"role_id": user.RoleID,
		"exp":     time.Now().Add(dur).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func buildAuthResponse(token string, user *models.User) *AuthResponse {
	perms := make([]PermissionResponse, 0, len(user.Role.RolePermissions))
	for _, rp := range user.Role.RolePermissions {
		perms = append(perms, PermissionResponse{
			ID:       rp.Permission.ID,
			Name:     rp.Permission.Name,
			Key:      rp.Permission.Key,
			Resource: rp.Permission.Resource,
		})
	}

	return &AuthResponse{
		AccessToken: token,
		User: UserResponse{
			ID:          user.ID,
			Name:        user.Name,
			Email:       user.Email,
			Avatar:      user.Avatar,
			PhoneNumber: user.PhoneNumber,
			Role: RoleResponse{
				ID:          user.Role.ID,
				Name:        user.Role.Name,
				Key:         user.Role.Key,
				Permissions: perms,
			},
		},
	}
}
