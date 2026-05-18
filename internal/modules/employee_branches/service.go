package employee_branches

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
	"kirimaja-go/models"
)

type Service interface {
	FindAll() ([]EmployeeBranchResponse, error)
	FindByID(id uint) (*EmployeeBranchResponse, error)
	Create(req CreateEmployeeBranchRequest) (*EmployeeBranchResponse, error)
	Update(id uint, req UpdateEmployeeBranchRequest) (*EmployeeBranchResponse, error)
	Delete(id uint) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo}
}

func (s *service) FindAll() ([]EmployeeBranchResponse, error) {
	list, err := s.repo.FindAll()
	if err != nil {
		return nil, err
	}
	result := make([]EmployeeBranchResponse, 0, len(list))
	for _, eb := range list {
		result = append(result, toResponse(eb))
	}
	return result, nil
}

func (s *service) FindByID(id uint) (*EmployeeBranchResponse, error) {
	eb, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if eb == nil {
		return nil, fmt.Errorf("employee branch with ID %d not found", id)
	}
	res := toResponse(*eb)
	return &res, nil
}

func (s *service) Create(req CreateEmployeeBranchRequest) (*EmployeeBranchResponse, error) {
	existing, err := s.repo.FindUserByEmail(req.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, fmt.Errorf("email %s is already in use", req.Email)
	}

	if branch, _ := s.repo.FindBranchByID(req.BranchID); branch == nil {
		return nil, fmt.Errorf("branch with ID %d does not exist", req.BranchID)
	}
	if role, _ := s.repo.FindRoleByID(req.RoleID); role == nil {
		return nil, fmt.Errorf("role with ID %d does not exist", req.RoleID)
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), 10)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	user := &models.User{
		Name:        req.Name,
		Email:       req.Email,
		PhoneNumber: req.PhoneNumber,
		Password:    string(hashed),
		Avatar:      req.Avatar,
		RoleID:      req.RoleID,
	}
	eb := &models.EmployeeBranch{BranchID: req.BranchID, Type: req.Type}

	if err := s.repo.CreateWithUser(user, eb); err != nil {
		return nil, err
	}
	return s.FindByID(eb.ID)
}

func (s *service) Update(id uint, req UpdateEmployeeBranchRequest) (*EmployeeBranchResponse, error) {
	existing, err := s.repo.FindByID(id)
	if err != nil || existing == nil {
		return nil, fmt.Errorf("employee branch with ID %d not found", id)
	}

	if req.Email != "" {
		user, _ := s.repo.FindUserByEmail(req.Email)
		if user != nil && user.ID != existing.UserID {
			return nil, fmt.Errorf("email %s is already in use", req.Email)
		}
	}
	if req.BranchID != 0 {
		if branch, _ := s.repo.FindBranchByID(req.BranchID); branch == nil {
			return nil, fmt.Errorf("branch with ID %d does not exist", req.BranchID)
		}
	}
	if req.RoleID != 0 {
		if role, _ := s.repo.FindRoleByID(req.RoleID); role == nil {
			return nil, fmt.Errorf("role with ID %d does not exist", req.RoleID)
		}
	}

	userData := map[string]any{}
	if req.Name != "" {
		userData["name"] = req.Name
	}
	if req.Email != "" {
		userData["email"] = req.Email
	}
	if req.PhoneNumber != "" {
		userData["phone_number"] = req.PhoneNumber
	}
	if req.Avatar != nil {
		userData["avatar"] = req.Avatar
	}
	if req.RoleID != 0 {
		userData["role_id"] = req.RoleID
	}
	if req.Password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), 10)
		if err != nil {
			return nil, errors.New("failed to hash password")
		}
		userData["password"] = string(hashed)
	}

	ebData := map[string]any{}
	if req.BranchID != 0 {
		ebData["branch_id"] = req.BranchID
	}
	if req.Type != "" {
		ebData["type"] = req.Type
	}

	if err := s.repo.UpdateWithUser(existing, userData, ebData); err != nil {
		return nil, err
	}
	return s.FindByID(id)
}

func (s *service) Delete(id uint) error {
	existing, err := s.repo.FindByID(id)
	if err != nil || existing == nil {
		return fmt.Errorf("employee branch with ID %d not found", id)
	}
	return s.repo.DeleteWithUser(existing)
}

func toResponse(eb models.EmployeeBranch) EmployeeBranchResponse {
	return EmployeeBranchResponse{
		ID:       eb.ID,
		UserID:   eb.UserID,
		BranchID: eb.BranchID,
		Type:     eb.Type,
		User: UserInfo{
			ID:          eb.User.ID,
			Name:        eb.User.Name,
			Email:       eb.User.Email,
			PhoneNumber: eb.User.PhoneNumber,
			Avatar:      eb.User.Avatar,
		},
		Branch: BranchInfo{
			ID:      eb.Branch.ID,
			Name:    eb.Branch.Name,
			Address: eb.Branch.Address,
		},
	}
}
