package models

import "time"

type ShipmentHistory struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	ShipmentID  uint      `gorm:"column:shipment_id" json:"shipment_id"`
	UserID      *uint     `gorm:"column:user_id" json:"user_id"`
	BranchID    *uint     `gorm:"column:branch_id" json:"branch_id"`
	Status      string    `json:"status"`
	Description *string   `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	Shipment Shipment `gorm:"foreignKey:ShipmentID" json:"shipment,omitempty"`
	User     *User    `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Branch   *Branch  `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
}

func (ShipmentHistory) TableName() string { return "shipment_histories" }
