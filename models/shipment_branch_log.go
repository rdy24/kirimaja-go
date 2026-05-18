package models

import "time"

type ShipmentBranchLog struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	ShipmentID      uint      `gorm:"column:shipment_id" json:"shipment_id"`
	BranchID        uint      `gorm:"column:branch_id" json:"branch_id"`
	TrackingNumber  string    `gorm:"column:tracking_number" json:"tracking_number"`
	Type            string    `json:"type"`
	Status          string    `json:"status"`
	Description     *string   `json:"description"`
	ScannedByUserID uint      `gorm:"column:scanned_by_user_id" json:"scanned_by_user_id"`
	ScanTime        time.Time `gorm:"column:scan_time" json:"scan_time"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`

	Shipment      Shipment `gorm:"foreignKey:ShipmentID" json:"shipment,omitempty"`
	Branch        Branch   `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	ScannedByUser User     `gorm:"foreignKey:ScannedByUserID" json:"scanned_by_user,omitempty"`
}

func (ShipmentBranchLog) TableName() string { return "shipment_branch_logs" }
