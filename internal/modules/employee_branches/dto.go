package employee_branches

type CreateEmployeeBranchRequest struct {
	Name        string  `json:"name" validate:"required"`
	Email       string  `json:"email" validate:"required,email"`
	PhoneNumber string  `json:"phone_number" validate:"required"`
	Password    string  `json:"password" validate:"required,min=8"`
	Avatar      *string `json:"avatar"`
	RoleID      uint    `json:"role_id" validate:"required"`
	BranchID    uint    `json:"branch_id" validate:"required"`
	Type        string  `json:"type" validate:"required"`
}

type UpdateEmployeeBranchRequest struct {
	Name        string  `json:"name"`
	Email       string  `json:"email" validate:"omitempty,email"`
	PhoneNumber string  `json:"phone_number"`
	Password    string  `json:"password" validate:"omitempty,min=8"`
	Avatar      *string `json:"avatar"`
	RoleID      uint    `json:"role_id"`
	BranchID    uint    `json:"branch_id"`
	Type        string  `json:"type"`
}

type UserInfo struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Email       string  `json:"email"`
	PhoneNumber string  `json:"phone_number"`
	Avatar      *string `json:"avatar"`
}

type BranchInfo struct {
	ID      uint   `json:"id"`
	Name    string `json:"name"`
	Address string `json:"address"`
}

type EmployeeBranchResponse struct {
	ID       uint       `json:"id"`
	UserID   uint       `json:"user_id"`
	BranchID uint       `json:"branch_id"`
	Type     string     `json:"type"`
	User     UserInfo   `json:"user"`
	Branch   BranchInfo `json:"branch"`
}
