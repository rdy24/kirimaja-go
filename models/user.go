package models

import "time"

type User struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `json:"name"`
	Email       string    `gorm:"uniqueIndex" json:"email"`
	Password    string    `json:"-"`
	Avatar      *string   `json:"avatar"`
	PhoneNumber string    `gorm:"column:phone_number" json:"phone_number"`
	RoleID      uint      `gorm:"column:role_id" json:"role_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	Role               Role                `gorm:"foreignKey:RoleID" json:"role,omitempty"`
	EmployeeBranches   []EmployeeBranch    `gorm:"foreignKey:UserID" json:"employee_branches,omitempty"`
	UserAddresses      []UserAddress       `gorm:"foreignKey:UserID" json:"user_addresses,omitempty"`
	ShipmentDetails    []ShipmentDetail    `gorm:"foreignKey:UserID" json:"shipment_details,omitempty"`
	ShipmentHistories  []ShipmentHistory   `gorm:"foreignKey:UserID" json:"shipment_histories,omitempty"`
	ShipmentBranchLogs []ShipmentBranchLog `gorm:"foreignKey:ScannedByUserID" json:"shipment_branch_logs,omitempty"`
}

func (User) TableName() string { return "users" }
