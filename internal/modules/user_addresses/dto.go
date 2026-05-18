package user_addresses

type CreateUserAddressRequest struct {
	Address string  `json:"address" form:"address" validate:"required"`
	Tag     string  `json:"tag" form:"tag" validate:"required"`
	Label   string  `json:"label" form:"label" validate:"required"`
	Photo   *string `json:"photo" form:"photo"`
}

type UpdateUserAddressRequest struct {
	Address string  `json:"address" form:"address"`
	Tag     string  `json:"tag" form:"tag"`
	Label   string  `json:"label" form:"label"`
	Photo   *string `json:"photo" form:"photo"`
}

type UserInfo struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Email       string  `json:"email"`
	PhoneNumber string  `json:"phone_number"`
	Avatar      *string `json:"avatar"`
}

type UserAddressResponse struct {
	ID        uint     `json:"id"`
	UserID    uint     `json:"user_id"`
	Address   string   `json:"address"`
	Tag       *string  `json:"tag"`
	Label     *string  `json:"label"`
	Photo     *string  `json:"photo"`
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
	User      UserInfo `json:"user"`
}
