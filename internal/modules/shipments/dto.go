package shipments

type CreateShipmentRequest struct {
	PickupAddressID    uint    `json:"pickup_address_id" validate:"required"`
	DestinationAddress string  `json:"destination_address" validate:"required"`
	RecipientName      string  `json:"recipient_name" validate:"required"`
	RecipientPhone     string  `json:"recipient_phone" validate:"required,min=10"`
	Weight             float64 `json:"weight" validate:"required,gt=0"`
	PackageType        string  `json:"package_type" validate:"required"`
	DeliveryType       string  `json:"delivery_type" validate:"required,oneof=same_day next_day regular"`
}

type ShipmentCost struct {
	TotalPrice    float64
	BasePrice     float64
	WeightPrice   float64
	DistancePrice float64
}

// WebhookPayload is the normalized Midtrans notification (transport concern,
// kept out of the service file).
type WebhookPayload struct {
	TransactionID     string
	TransactionStatus string
	OrderID           string
	GrossAmount       string
	StatusCode        string
	SignatureKey      string
	PaymentType       string
}

type ScanShipmentRequest struct {
	TrackingNumber  string `json:"tracking_number" binding:"required"`
	Type            string `json:"type" binding:"required,oneof=IN OUT"`
	IsReadyToPickup bool   `json:"is_ready_to_pickup"`
}
