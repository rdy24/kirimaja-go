package models

import "time"

type EmployeeBranch struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"column:user_id" json:"user_id"`
	BranchID  uint      `gorm:"column:branch_id" json:"branch_id"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	User   User   `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Branch Branch `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
}

func (EmployeeBranch) TableName() string { return "employee_branches" }
