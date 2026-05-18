package shipments

import (
	"time"

	"kirimaja-go/models"
)

// Response DTOs decouple the HTTP contract from the GORM models. Unlike the
// raw models (value-type relations with ineffective `omitempty` that leak
// zero-filled `{}` objects and the whole relation graph), these expose only
// what the API intends, and nested relations are pointers so an unloaded
// relation is omitted entirely.
//
// CONTRACT CHANGE vs. the old model-serialized responses:
//   - zero-valued nested objects (e.g. "user":{"id":0,...}, "role":{},
//     "shipment":{...}) are no longer emitted; absent relations are omitted.
//   - the recursive shipment_detail.shipment / payment.shipment echo is gone.
//   - user objects no longer carry their role/permission sub-graph here.

type UserResponse struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Email       string  `json:"email"`
	Avatar      *string `json:"avatar"`
	PhoneNumber string  `json:"phone_number"`
	RoleID      uint    `json:"role_id"`
}

type AddressResponse struct {
	ID        uint     `json:"id"`
	UserID    uint     `json:"user_id"`
	Address   string   `json:"address"`
	Tag       *string  `json:"tag"`
	Label     *string  `json:"label"`
	Photo     *string  `json:"photo"`
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
}

type ShipmentDetailResponse struct {
	ID                   uint             `json:"id"`
	ShipmentID           uint             `json:"shipment_id"`
	UserID               uint             `json:"user_id"`
	PickupAddressID      uint             `json:"pickup_address_id"`
	Weight               *float64         `json:"weight"`
	DeliveryType         string           `json:"delivery_type"`
	DestinationAddress   string           `json:"destination_address"`
	DestinationLatitude  *float64         `json:"destination_latitude"`
	DestinationLongitude *float64         `json:"destination_longitude"`
	PackageType          string           `json:"package_type"`
	PickupProof          *string          `json:"pickup_proof"`
	ReceiptProof         *string          `json:"receipt_proof"`
	RecipientName        string           `json:"recipient_name"`
	RecipientPhone       string           `json:"recipient_phone"`
	BasePrice            *float64         `json:"base_price"`
	WeightPrice          *float64         `json:"weight_price"`
	DistancePrice        *float64         `json:"distance_price"`
	User                 *UserResponse    `json:"user,omitempty"`
	Address              *AddressResponse `json:"address,omitempty"`
}

type PaymentResponse struct {
	ID            uint       `json:"id"`
	ShipmentID    uint       `json:"shipment_id"`
	ExternalID    *string    `json:"external_id"`
	InvoiceID     *string    `json:"invoice_id"`
	PaymentMethod *string    `json:"payment_method"`
	Status        *string    `json:"status"`
	InvoiceUrl    *string    `json:"invoice_url"`
	ExpiryDate    *time.Time `json:"expiry_date"`
}

type ShipmentHistoryResponse struct {
	ID          uint      `json:"id"`
	ShipmentID  uint      `json:"shipment_id"`
	UserID      *uint     `json:"user_id"`
	BranchID    *uint     `json:"branch_id"`
	Status      string    `json:"status"`
	Description *string   `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type BranchResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Address     string `json:"address"`
	PhoneNumber string `json:"phone_number"`
}

type ShipmentResponse struct {
	ID                uint                      `json:"id"`
	PaymentStatus     string                    `json:"payment_status"`
	DeliveryStatus    *string                   `json:"delivery_status"`
	TrackingNumber    *string                   `json:"tracking_number"`
	QrCodeImage       *string                   `json:"qr_code_image"`
	Price             *float64                  `json:"price"`
	Distance          *float64                  `json:"distance"`
	CurrentBranchID   *uint                     `json:"current_branch_id"`
	CreatedAt         time.Time                 `json:"created_at"`
	UpdatedAt         time.Time                 `json:"updated_at"`
	ShipmentDetail    *ShipmentDetailResponse   `json:"shipment_detail,omitempty"`
	Payment           *PaymentResponse          `json:"payment,omitempty"`
	ShipmentHistories []ShipmentHistoryResponse `json:"shipment_histories,omitempty"`
}

type ShipmentBranchLogResponse struct {
	ID              uint              `json:"id"`
	ShipmentID      uint              `json:"shipment_id"`
	BranchID        uint              `json:"branch_id"`
	TrackingNumber  string            `json:"tracking_number"`
	Type            string            `json:"type"`
	Status          string            `json:"status"`
	Description     *string           `json:"description"`
	ScannedByUserID uint              `json:"scanned_by_user_id"`
	ScanTime        time.Time         `json:"scan_time"`
	CreatedAt       time.Time         `json:"created_at"`
	Shipment        *ShipmentResponse `json:"shipment,omitempty"`
	Branch          *BranchResponse   `json:"branch,omitempty"`
	ScannedByUser   *UserResponse     `json:"scanned_by_user,omitempty"`
}

// --- mappers (all nil-safe) ------------------------------------------------

func toUserResponse(u *models.User) *UserResponse {
	if u == nil || u.ID == 0 {
		return nil
	}
	return &UserResponse{
		ID: u.ID, Name: u.Name, Email: u.Email,
		Avatar: u.Avatar, PhoneNumber: u.PhoneNumber, RoleID: u.RoleID,
	}
}

func toAddressResponse(a *models.UserAddress) *AddressResponse {
	if a == nil || a.ID == 0 {
		return nil
	}
	return &AddressResponse{
		ID: a.ID, UserID: a.UserID, Address: a.Address, Tag: a.Tag,
		Label: a.Label, Photo: a.Photo, Latitude: a.Latitude, Longitude: a.Longitude,
	}
}

func toShipmentDetailResponse(d *models.ShipmentDetail) *ShipmentDetailResponse {
	if d == nil {
		return nil
	}
	return &ShipmentDetailResponse{
		ID: d.ID, ShipmentID: d.ShipmentID, UserID: d.UserID,
		PickupAddressID: d.PickupAddressID, Weight: d.Weight, DeliveryType: d.DeliveryType,
		DestinationAddress: d.DestinationAddress, DestinationLatitude: d.DestinationLatitude,
		DestinationLongitude: d.DestinationLongitude, PackageType: d.PackageType,
		PickupProof: d.PickupProof, ReceiptProof: d.ReceiptProof,
		RecipientName: d.RecipientName, RecipientPhone: d.RecipientPhone,
		BasePrice: d.BasePrice, WeightPrice: d.WeightPrice, DistancePrice: d.DistancePrice,
		User:    toUserResponse(&d.User),
		Address: toAddressResponse(&d.Address),
	}
}

func toPaymentResponse(p *models.Payment) *PaymentResponse {
	if p == nil || p.ID == 0 {
		return nil
	}
	return &PaymentResponse{
		ID: p.ID, ShipmentID: p.ShipmentID, ExternalID: p.ExternalID,
		InvoiceID: p.InvoiceID, PaymentMethod: p.PaymentMethod, Status: p.Status,
		InvoiceUrl: p.InvoiceUrl, ExpiryDate: p.ExpiryDate,
	}
}

func toHistoryResponses(hs []models.ShipmentHistory) []ShipmentHistoryResponse {
	if len(hs) == 0 {
		return nil
	}
	out := make([]ShipmentHistoryResponse, 0, len(hs))
	for _, h := range hs {
		out = append(out, ShipmentHistoryResponse{
			ID: h.ID, ShipmentID: h.ShipmentID, UserID: h.UserID, BranchID: h.BranchID,
			Status: h.Status, Description: h.Description, CreatedAt: h.CreatedAt,
		})
	}
	return out
}

func toBranchResponse(b *models.Branch) *BranchResponse {
	if b == nil || b.ID == 0 {
		return nil
	}
	return &BranchResponse{ID: b.ID, Name: b.Name, Address: b.Address, PhoneNumber: b.PhoneNumber}
}

func ToShipmentResponse(s *models.Shipment) *ShipmentResponse {
	if s == nil {
		return nil
	}
	return &ShipmentResponse{
		ID: s.ID, PaymentStatus: s.PaymentStatus, DeliveryStatus: s.DeliveryStatus,
		TrackingNumber: s.TrackingNumber, QrCodeImage: s.QrCodeImage, Price: s.Price,
		Distance: s.Distance, CurrentBranchID: s.CurrentBranchID,
		CreatedAt: s.CreatedAt, UpdatedAt: s.UpdatedAt,
		ShipmentDetail:    toShipmentDetailResponse(s.ShipmentDetail),
		Payment:           toPaymentResponse(s.Payment),
		ShipmentHistories: toHistoryResponses(s.ShipmentHistories),
	}
}

func ToShipmentResponses(list []models.Shipment) []ShipmentResponse {
	out := make([]ShipmentResponse, 0, len(list))
	for i := range list {
		out = append(out, *ToShipmentResponse(&list[i]))
	}
	return out
}

func ToBranchLogResponse(l *models.ShipmentBranchLog) *ShipmentBranchLogResponse {
	if l == nil {
		return nil
	}
	r := &ShipmentBranchLogResponse{
		ID: l.ID, ShipmentID: l.ShipmentID, BranchID: l.BranchID,
		TrackingNumber: l.TrackingNumber, Type: l.Type, Status: l.Status,
		Description: l.Description, ScannedByUserID: l.ScannedByUserID,
		ScanTime: l.ScanTime, CreatedAt: l.CreatedAt,
		Branch:        toBranchResponse(&l.Branch),
		ScannedByUser: toUserResponse(&l.ScannedByUser),
	}
	if l.Shipment.ID != 0 {
		r.Shipment = ToShipmentResponse(&l.Shipment)
	}
	return r
}

func ToBranchLogResponses(list []models.ShipmentBranchLog) []ShipmentBranchLogResponse {
	out := make([]ShipmentBranchLogResponse, 0, len(list))
	for i := range list {
		out = append(out, *ToBranchLogResponse(&list[i]))
	}
	return out
}
