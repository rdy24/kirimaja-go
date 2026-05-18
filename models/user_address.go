package models

import "time"

type UserAddress struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"column:user_id" json:"user_id"`
	Address   string    `json:"address"`
	Tag       *string   `json:"tag"`
	Label     *string   `json:"label"`
	Photo     *string   `json:"photo"`
	Latitude  *float64  `json:"latitude"`
	Longitude *float64  `json:"longitude"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	User            User             `gorm:"foreignKey:UserID" json:"user,omitempty"`
	ShipmentDetails []ShipmentDetail `gorm:"foreignKey:PickupAddressID" json:"shipment_details,omitempty"`
}

func (UserAddress) TableName() string { return "user_addresses" }
