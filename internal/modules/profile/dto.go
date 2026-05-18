package profile

type UpdateProfileRequest struct {
	Name        string `json:"name"`
	Email       string `json:"email" validate:"omitempty,email"`
	PhoneNumber string `json:"phone_number"`
	Password    string `json:"password" validate:"omitempty,min=8"`
}

type ProfileResponse struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Email       string  `json:"email"`
	Avatar      *string `json:"avatar"`
	PhoneNumber string  `json:"phone_number"`
}
