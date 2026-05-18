package auth

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type RegisterRequest struct {
	Name        string `json:"name" validate:"required,min=1"`
	Email       string `json:"email" validate:"required,email"`
	Password    string `json:"password" validate:"required,min=8"`
	PhoneNumber string `json:"phone_number" validate:"required,min=10"`
}

type PermissionResponse struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Key      string `json:"key"`
	Resource string `json:"resource"`
}

type RoleResponse struct {
	ID          uint                 `json:"id"`
	Name        string               `json:"name"`
	Key         string               `json:"key"`
	Permissions []PermissionResponse `json:"permissions"`
}

type UserResponse struct {
	ID          uint         `json:"id"`
	Name        string       `json:"name"`
	Email       string       `json:"email"`
	Avatar      *string      `json:"avatar"`
	PhoneNumber string       `json:"phone_number"`
	Role        RoleResponse `json:"role"`
}

type AuthResponse struct {
	AccessToken string       `json:"access_token"`
	User        UserResponse `json:"user"`
}
