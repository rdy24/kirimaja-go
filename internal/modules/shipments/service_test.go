package shipments

import (
	"context"
	"errors"
	"testing"
	"time"

	"kirimaja-go/internal/common/midtrans"
	"kirimaja-go/internal/common/opencage"
	"kirimaja-go/internal/common/worker"
	"kirimaja-go/models"
)

// --- fakes -----------------------------------------------------------------

type fakeRepo struct {
	shipmentByID       map[uint]*models.Shipment
	shipmentByTracking map[string]*models.Shipment
	pickupAddr         *models.UserAddress
	paymentByExt       *models.Payment
	employeeBranch     *models.EmployeeBranch
	lastBranchLogIn    *models.ShipmentBranchLog

	createdShipments []*models.Shipment
	createdPayments  []*models.Payment
	createdHistory   []*models.ShipmentHistory
	shipmentUpdates  []map[string]any
	atomicCalls      int
}

func (f *fakeRepo) Atomic(_ context.Context, fn func(Repository) error) error {
	f.atomicCalls++
	return fn(f)
}
func (f *fakeRepo) FindAll(context.Context, uint) ([]models.Shipment, error) { return nil, nil }
func (f *fakeRepo) FindByID(_ context.Context, id uint) (*models.Shipment, error) {
	return f.shipmentByID[id], nil
}
func (f *fakeRepo) FindByTrackingNumber(_ context.Context, t string) (*models.Shipment, error) {
	return f.shipmentByTracking[t], nil
}
func (f *fakeRepo) FindPickupAddress(context.Context, uint) (*models.UserAddress, error) {
	return f.pickupAddr, nil
}
func (f *fakeRepo) CreateShipment(_ context.Context, s *models.Shipment) error {
	s.ID = 1
	f.createdShipments = append(f.createdShipments, s)
	return nil
}
func (f *fakeRepo) CreateShipmentDetail(context.Context, *models.ShipmentDetail) error { return nil }
func (f *fakeRepo) CreatePayment(_ context.Context, p *models.Payment) error {
	p.ID = 1
	f.createdPayments = append(f.createdPayments, p)
	return nil
}
func (f *fakeRepo) CreateHistory(_ context.Context, h *models.ShipmentHistory) error {
	f.createdHistory = append(f.createdHistory, h)
	return nil
}
func (f *fakeRepo) UpdateShipment(_ context.Context, _ uint, data map[string]any) error {
	f.shipmentUpdates = append(f.shipmentUpdates, data)
	return nil
}
func (f *fakeRepo) UpdatePayment(context.Context, uint, map[string]any) error        { return nil }
func (f *fakeRepo) UpdateShipmentDetail(context.Context, uint, map[string]any) error { return nil }
func (f *fakeRepo) FindPaymentByExternalID(context.Context, string) (*models.Payment, error) {
	return f.paymentByExt, nil
}
func (f *fakeRepo) FindEmployeeBranch(context.Context, uint) (*models.EmployeeBranch, error) {
	return f.employeeBranch, nil
}
func (f *fakeRepo) FindAllBranchLogs(context.Context, *uint) ([]models.ShipmentBranchLog, error) {
	return nil, nil
}
func (f *fakeRepo) CreateBranchLog(_ context.Context, l *models.ShipmentBranchLog) (*models.ShipmentBranchLog, error) {
	l.ID = 1
	return l, nil
}
func (f *fakeRepo) FindLastBranchLogIn(context.Context, string, uint) (*models.ShipmentBranchLog, error) {
	return f.lastBranchLogIn, nil
}
func (f *fakeRepo) FindAllForCourier(context.Context) ([]models.Shipment, error) { return nil, nil }

type fakeGeo struct{ err error }

func (f fakeGeo) GeocodeContext(context.Context, string) (*opencage.Location, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &opencage.Location{Lat: 1, Lng: 2}, nil
}

type fakeGateway struct {
	snapErr  error
	verifyOK bool
}

func (f fakeGateway) CreateSnap(string, int64, string) (*midtrans.SnapResult, error) {
	if f.snapErr != nil {
		return nil, f.snapErr
	}
	return &midtrans.SnapResult{RedirectURL: "https://pay.example/x"}, nil
}
func (f fakeGateway) VerifyWebhookSignature(string, string, string) func(string) bool {
	return func(string) bool { return f.verifyOK }
}

type fakeQR struct{ calls int }

func (f *fakeQR) Generate(string) (string, error) { f.calls++; return "/qrcodes/x.png", nil }

type fakeTask struct{ successCalls int }

func (f *fakeTask) EnqueuePaymentNotification(worker.PaymentNotificationPayload) error { return nil }
func (f *fakeTask) EnqueuePaymentSuccess(worker.PaymentSuccessPayload) error {
	f.successCalls++
	return nil
}
func (f *fakeTask) EnqueuePaymentExpiry(worker.PaymentExpiryPayload, time.Time) error { return nil }
func (f *fakeTask) CancelPaymentExpiry(uint)                                          {}

// --- helpers ---------------------------------------------------------------

func newSvc(repo Repository, deps ...any) *service {
	s := &service{repo: repo}
	for _, d := range deps {
		switch v := d.(type) {
		case Geocoder:
			s.geocli = v
		case PaymentGateway:
			s.midtrans = v
		case QRGenerator:
			s.qrSvc = v
		case TaskQueue:
			s.worker = v
		}
	}
	return s
}

// --- tests -----------------------------------------------------------------

func TestFindByID_Authorization(t *testing.T) {
	repo := &fakeRepo{shipmentByID: map[uint]*models.Shipment{
		1: {ID: 1, ShipmentDetail: &models.ShipmentDetail{UserID: 7}},
	}}
	s := newSvc(repo)
	ctx := context.Background()

	if _, err := s.FindByID(ctx, 1, 7, 2); err != nil {
		t.Fatalf("owner should be allowed, got %v", err)
	}
	if _, err := s.FindByID(ctx, 1, 99, 2); !errors.Is(err, ErrForbidden) {
		t.Fatalf("non-owner must get ErrForbidden, got %v", err)
	}
	if _, err := s.FindByID(ctx, 1, 99, superAdminRoleID); err != nil {
		t.Fatalf("super admin should be allowed, got %v", err)
	}
}

func TestHandleWebhook_Idempotent(t *testing.T) {
	pending := StatusPending
	pay := &models.Payment{
		ID:     1,
		Status: &pending,
		Shipment: models.Shipment{
			ID:             1,
			ShipmentDetail: &models.ShipmentDetail{User: models.User{Email: "a@b.c"}},
		},
	}
	repo := &fakeRepo{paymentByExt: pay}
	qr := &fakeQR{}
	task := &fakeTask{}
	s := newSvc(repo, fakeGateway{verifyOK: true}, QRGenerator(qr), TaskQueue(task))
	ctx := context.Background()

	p := WebhookPayload{OrderID: "INV-1", TransactionStatus: "settlement", TransactionID: "TX1"}
	if err := s.HandleWebhook(ctx, p); err != nil {
		t.Fatalf("first webhook: %v", err)
	}
	if qr.calls != 1 || task.successCalls != 1 {
		t.Fatalf("first settlement should generate QR + email once, got qr=%d email=%d", qr.calls, task.successCalls)
	}

	// Simulate the now-persisted PAID state, then replay (Midtrans retries).
	paid := StatusPaid
	pay.Status = &paid
	if err := s.HandleWebhook(ctx, p); err != nil {
		t.Fatalf("replay webhook: %v", err)
	}
	if qr.calls != 1 || task.successCalls != 1 {
		t.Fatalf("replay must be a no-op, got qr=%d email=%d", qr.calls, task.successCalls)
	}
}

func TestCreate_NoOrphanWhenGatewayFails(t *testing.T) {
	lat, lng := 1.0, 2.0
	repo := &fakeRepo{pickupAddr: &models.UserAddress{
		UserID: 7, Latitude: &lat, Longitude: &lng, User: models.User{Email: "a@b.c"},
	}}
	s := newSvc(repo, fakeGeo{}, fakeGateway{snapErr: errors.New("gateway down")})

	_, err := s.Create(context.Background(), 7, CreateShipmentRequest{Weight: 1000, DeliveryType: "regular"})
	if err == nil {
		t.Fatal("expected error when gateway fails")
	}
	if len(repo.createdShipments) != 0 || repo.atomicCalls != 0 {
		t.Fatalf("no DB writes must happen if gateway fails (no orphan): shipments=%d atomic=%d",
			len(repo.createdShipments), repo.atomicCalls)
	}
}

func TestPickShipment_BranchClaimAndScoping(t *testing.T) {
	ctx := context.Background()

	// Unclaimed shipment → first courier claims it to their branch.
	repo := &fakeRepo{
		shipmentByTracking: map[string]*models.Shipment{"KA1": {ID: 1}},
		employeeBranch:     &models.EmployeeBranch{BranchID: 5},
	}
	s := newSvc(repo)
	if _, err := s.PickShipment(ctx, "KA1", 7); err != nil {
		t.Fatalf("claim should succeed: %v", err)
	}
	claimed := false
	for _, u := range repo.shipmentUpdates {
		if v, ok := u["current_branch_id"]; ok && v == uint(5) {
			claimed = true
		}
	}
	if !claimed {
		t.Fatal("unclaimed shipment must be claimed to courier's branch")
	}

	// Shipment owned by another branch → courier from branch 5 is rejected.
	other := uint(9)
	repo2 := &fakeRepo{
		shipmentByTracking: map[string]*models.Shipment{"KA2": {ID: 2, CurrentBranchID: &other}},
		employeeBranch:     &models.EmployeeBranch{BranchID: 5},
	}
	s2 := newSvc(repo2)
	if _, err := s2.PickShipment(ctx, "KA2", 7); !errors.Is(err, ErrForbidden) {
		t.Fatalf("courier from another branch must get ErrForbidden, got %v", err)
	}
}

func inTransit() *string { s := StatusInTransit; return &s }

func TestScanShipment_RejectsInvalidStatus(t *testing.T) {
	ready := StatusReadyToPickup
	repo := &fakeRepo{
		shipmentByTracking: map[string]*models.Shipment{"KA1": {ID: 1, DeliveryStatus: &ready}},
		employeeBranch:     &models.EmployeeBranch{BranchID: 5},
	}
	s := newSvc(repo)
	_, err := s.ScanShipment(context.Background(), ScanShipmentRequest{TrackingNumber: "KA1", Type: "IN"}, 7)
	if err == nil {
		t.Fatal("scanning a shipment not in a scannable status must fail")
	}
}

func TestScanShipment_InTakesBranchOwnership(t *testing.T) {
	repo := &fakeRepo{
		shipmentByTracking: map[string]*models.Shipment{"KA1": {ID: 1, DeliveryStatus: inTransit()}},
		employeeBranch:     &models.EmployeeBranch{BranchID: 5, Branch: models.Branch{Name: "Hub"}},
	}
	s := newSvc(repo)
	log, err := s.ScanShipment(context.Background(), ScanShipmentRequest{TrackingNumber: "KA1", Type: "IN"}, 7)
	if err != nil {
		t.Fatalf("valid IN scan failed: %v", err)
	}
	if log == nil {
		t.Fatal("expected a branch log")
	}
	owned := false
	for _, u := range repo.shipmentUpdates {
		if v, ok := u["current_branch_id"]; ok && v == uint(5) {
			if u["delivery_status"] != StatusArrivedAtBranch {
				t.Fatalf("IN scan should set ARRIVED_AT_BRANCH, got %v", u["delivery_status"])
			}
			owned = true
		}
	}
	if !owned {
		t.Fatal("scanning branch must take ownership of the shipment")
	}
}

func TestScanShipment_OutRequiresPriorIn(t *testing.T) {
	repo := &fakeRepo{
		shipmentByTracking: map[string]*models.Shipment{"KA1": {ID: 1, DeliveryStatus: inTransit()}},
		employeeBranch:     &models.EmployeeBranch{BranchID: 5, Branch: models.Branch{Name: "Hub"}},
		lastBranchLogIn:    nil, // no prior IN at this branch
	}
	s := newSvc(repo)
	_, err := s.ScanShipment(context.Background(), ScanShipmentRequest{TrackingNumber: "KA1", Type: "OUT"}, 7)
	if err == nil {
		t.Fatal("OUT scan without a prior IN at the same branch must fail")
	}
}

func TestCourierActions_RejectForeignBranch(t *testing.T) {
	ctx := context.Background()
	other := uint(9)
	mk := func() *fakeRepo {
		return &fakeRepo{
			shipmentByTracking: map[string]*models.Shipment{"KA1": {ID: 1, CurrentBranchID: &other}},
			employeeBranch:     &models.EmployeeBranch{BranchID: 5},
		}
	}
	if _, err := newSvc(mk()).PickupShipment(ctx, "KA1", 7, "p.jpg"); !errors.Is(err, ErrForbidden) {
		t.Fatalf("PickupShipment from foreign branch must be ErrForbidden, got %v", err)
	}
	if _, err := newSvc(mk()).DeliverToCustomer(ctx, "KA1", 7, "p.jpg"); !errors.Is(err, ErrForbidden) {
		t.Fatalf("DeliverToCustomer from foreign branch must be ErrForbidden, got %v", err)
	}
	if _, err := newSvc(mk()).DeliverToBranch(ctx, "KA1", 7); !errors.Is(err, ErrForbidden) {
		t.Fatalf("DeliverToBranch from foreign branch must be ErrForbidden, got %v", err)
	}
}
