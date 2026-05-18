package models

import "time"

type ShipmentDetail struct {
	ID                   uint      `gorm:"primaryKey" json:"id"`
	ShipmentID           uint      `gorm:"column:shipment_id;uniqueIndex" json:"shipment_id"`
	UserID               uint      `gorm:"column:user_id" json:"user_id"`
	PickupAddressID      uint      `gorm:"column:pickup_address_id" json:"pickup_address_id"`
	Weight               *float64  `json:"weight"`
	DeliveryType         string    `gorm:"column:delivery_type" json:"delivery_type"`
	DestinationAddress   string    `gorm:"column:destination_address" json:"destination_address"`
	DestinationLatitude  *float64  `gorm:"column:destination_latitude" json:"destination_latitude"`
	DestinationLongitude *float64  `gorm:"column:destination_longitude" json:"destination_longitude"`
	PackageType          string    `gorm:"column:package_type" json:"package_type"`
	PickupProof          *string   `gorm:"column:pickup_proof" json:"pickup_proof"`
	ReceiptProof         *string   `gorm:"column:receipt_proof" json:"receipt_proof"`
	RecipientName        string    `gorm:"column:recipient_name" json:"recipient_name"`
	RecipientPhone       string    `gorm:"column:recipient_phone" json:"recipient_phone"`
	BasePrice            *float64  `gorm:"column:base_price" json:"base_price"`
	WeightPrice          *float64  `gorm:"column:weight_price" json:"weight_price"`
	DistancePrice        *float64  `gorm:"column:distance_price" json:"distance_price"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`

	Shipment Shipment    `gorm:"foreignKey:ShipmentID" json:"shipment,omitempty"`
	User     User        `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Address  UserAddress `gorm:"foreignKey:PickupAddressID" json:"address,omitempty"`
}

func (ShipmentDetail) TableName() string { return "shipment_details" }
