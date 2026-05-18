package permissions

type CreatePermissionRequest struct {
	Name     string `json:"name" validate:"required"`
	Key      string `json:"key" validate:"required"`
	Resource string `json:"resource" validate:"required"`
}

type PermissionResponse struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Key      string `json:"key"`
	Resource string `json:"resource"`
}
