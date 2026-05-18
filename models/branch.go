package models

import "time"

type Branch struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `json:"name"`
	Address     string    `json:"address"`
	PhoneNumber string    `gorm:"column:phone_number" json:"phone_number"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	EmployeeBranches  []EmployeeBranch   `gorm:"foreignKey:BranchID" json:"employee_branches,omitempty"`
	ShipmentHistories []ShipmentHistory  `gorm:"foreignKey:BranchID" json:"shipment_histories,omitempty"`
	ShipmentBranchLogs []ShipmentBranchLog `gorm:"foreignKey:BranchID" json:"shipment_branch_logs,omitempty"`
}

func (Branch) TableName() string { return "branches" }
