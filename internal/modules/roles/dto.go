package roles

type CreateRoleRequest struct {
	Name string `json:"name" validate:"required"`
	Key  string `json:"key" validate:"required"`
}

type UpdateRoleRequest struct {
	PermissionIDs []uint `json:"permission_ids" validate:"required"`
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
