package shipments

import (
	"encoding/json"
	"strings"
	"testing"

	"kirimaja-go/models"
)

func TestToShipmentResponse_NilSafe(t *testing.T) {
	if ToShipmentResponse(nil) != nil {
		t.Fatal("nil shipment must map to nil")
	}
}

func TestToShipmentResponse_OmitsUnloadedRelations(t *testing.T) {
	// Bare shipment (Create-style): no detail/payment/histories loaded.
	s := &models.Shipment{ID: 1, PaymentStatus: StatusPending}
	b, _ := json.Marshal(ToShipmentResponse(s))
	js := string(b)
	for _, k := range []string{"shipment_detail", "payment", "shipment_histories"} {
		if strings.Contains(js, `"`+k+`":`) {
			t.Fatalf("unloaded relation %q must be omitted, got: %s", k, js)
		}
	}
}

func TestToShipmentResponse_NoZeroJunkOrPasswordLeak(t *testing.T) {
	// Detail loaded, but its User relation NOT loaded (zero value). The old
	// model serialization emitted "user":{"id":0,...}; the DTO must omit it.
	s := &models.Shipment{
		ID: 1, PaymentStatus: StatusPaid,
		ShipmentDetail: &models.ShipmentDetail{ID: 9, ShipmentID: 1, RecipientName: "Budi"},
	}
	b, _ := json.Marshal(ToShipmentResponse(s))
	js := string(b)
	if strings.Contains(js, `"user"`) {
		t.Fatalf("zero User must be omitted, got: %s", js)
	}
	if strings.Contains(js, "password") || strings.Contains(js, "Password") {
		t.Fatalf("password must never appear, got: %s", js)
	}

	// Now with a real User loaded → it appears, still no password.
	s.ShipmentDetail.User = models.User{ID: 5, Email: "a@b.c", Password: "secret-hash"}
	b, _ = json.Marshal(ToShipmentResponse(s))
	js = string(b)
	if !strings.Contains(js, `"email":"a@b.c"`) {
		t.Fatalf("loaded user should be present, got: %s", js)
	}
	if strings.Contains(js, "secret-hash") {
		t.Fatalf("password hash leaked: %s", js)
	}
}

func TestToBranchLogResponse_NilSafe(t *testing.T) {
	if ToBranchLogResponse(nil) != nil {
		t.Fatal("nil branch log must map to nil")
	}
	l := &models.ShipmentBranchLog{ID: 1, BranchID: 2, TrackingNumber: "KA1"}
	b, _ := json.Marshal(ToBranchLogResponse(l))
	if strings.Contains(string(b), `"shipment"`) {
		t.Fatalf("unloaded nested shipment must be omitted, got: %s", string(b))
	}
}
