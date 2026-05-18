package models

import "time"

type Payment struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	ShipmentID    uint       `gorm:"column:shipment_id;uniqueIndex" json:"shipment_id"`
	ExternalID    *string    `gorm:"column:external_id" json:"external_id"`
	InvoiceID     *string    `gorm:"column:invoice_id" json:"invoice_id"`
	PaymentMethod *string    `gorm:"column:payment_method" json:"payment_method"`
	Status        *string    `json:"status"`
	InvoiceUrl    *string    `gorm:"column:invoice_url" json:"invoice_url"`
	ExpiryDate    *time.Time `gorm:"column:expiry_date" json:"expiry_date"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`

	Shipment Shipment `gorm:"foreignKey:ShipmentID" json:"shipment,omitempty"`
}

func (Payment) TableName() string { return "payments" }
