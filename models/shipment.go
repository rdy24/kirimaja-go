package models

import "time"

type Shipment struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	PaymentStatus  string    `gorm:"column:payment_status" json:"payment_status"`
	DeliveryStatus *string   `gorm:"column:delivery_status" json:"delivery_status"`
	TrackingNumber *string   `gorm:"column:tracking_number" json:"tracking_number"`
	QrCodeImage    *string   `gorm:"column:qr_code_image" json:"qr_code_image"`
	Price          *float64  `json:"price"`
	Distance       *float64  `json:"distance"`
	// CurrentBranchID is the branch that currently "owns" the shipment for
	// courier operations. NULL = not yet claimed (just paid); set on first
	// courier pickup and updated by each branch scan. Used to stop a courier
	// from another branch processing a shipment that isn't theirs.
	CurrentBranchID *uint     `gorm:"column:current_branch_id;index" json:"current_branch_id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`

	ShipmentDetail     *ShipmentDetail      `gorm:"foreignKey:ShipmentID" json:"shipment_detail,omitempty"`
	ShipmentHistories  []ShipmentHistory    `gorm:"foreignKey:ShipmentID" json:"shipment_histories,omitempty"`
	Payment            *Payment             `gorm:"foreignKey:ShipmentID" json:"payment,omitempty"`
	ShipmentBranchLogs []ShipmentBranchLog  `gorm:"foreignKey:ShipmentID" json:"shipment_branch_logs,omitempty"`
}

func (Shipment) TableName() string { return "shipments" }
