package pdf

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"html/template"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"

	"kirimaja-go/internal/common/storage"
)

type ShipmentData struct {
	TrackingNumber     string
	ShipmentID         uint
	CreatedDate        string
	DeliveryType       string
	PackageType        string
	Weight             string
	Price              string
	Distance           string
	PaymentStatus      string
	DeliveryStatus     string
	BasePrice          string
	WeightPrice        string
	DistancePrice      string
	SenderName         string
	SenderEmail        string
	SenderPhone        string
	PickupAddress      string
	RecipientName      string
	RecipientPhone     string
	DestinationAddress string
	QRCodeBase64       string
	GeneratedDate      string
}

// maxConcurrentPDF caps how many headless-Chrome renders run at once. PDF
// generation spawns a Chrome instance; without a ceiling a burst of requests
// exhausts memory (effectively a DoS).
const maxConcurrentPDF = 4

type Service struct {
	allocCtx    context.Context
	allocCancel context.CancelFunc
	store       storage.Store
	tmpl        *template.Template
	sem         chan struct{}
}

func New(store storage.Store) (*Service, func()) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-setuid-sandbox", true),
	)
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	svc := &Service{
		allocCtx:    allocCtx,
		allocCancel: cancel,
		store:       store,
		tmpl:        template.Must(template.New("shipping").Parse(htmlTemplate)),
		sem:         make(chan struct{}, maxConcurrentPDF),
	}
	return svc, cancel
}

func (s *Service) Generate(ctx context.Context, data ShipmentData) ([]byte, error) {
	// Acquire a render slot; give up immediately if the request is already
	// gone instead of queueing work nobody is waiting for.
	select {
	case s.sem <- struct{}{}:
		defer func() { <-s.sem }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	var buf bytes.Buffer
	if err := s.tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("render template: %w", err)
	}
	htmlContent := buf.String()

	chromeCtx, cancel := chromedp.NewContext(s.allocCtx)
	defer cancel()
	// Free the Chrome instance early if the originating request is cancelled.
	defer context.AfterFunc(ctx, cancel)()
	chromeCtx, cancelTimeout := context.WithTimeout(chromeCtx, 30*time.Second)
	defer cancelTimeout()

	var pdfBuf []byte
	err := chromedp.Run(chromeCtx,
		chromedp.Navigate("about:blank"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			frameTree, err := page.GetFrameTree().Do(ctx)
			if err != nil {
				return err
			}
			return page.SetDocumentContent(frameTree.Frame.ID, htmlContent).Do(ctx)
		}),
		chromedp.WaitReady("body"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			buf, _, err := page.PrintToPDF().
				WithPrintBackground(true).
				WithMarginTop(0.4).
				WithMarginBottom(0.4).
				WithMarginLeft(0.2).
				WithMarginRight(0.2).
				Do(ctx)
			pdfBuf = buf
			return err
		}),
	)
	return pdfBuf, err
}

func (s *Service) QRBase64(ctx context.Context, key string) string {
	if key == "" {
		return ""
	}
	data, err := s.store.Get(ctx, key)
	if err != nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(data)
}

const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"/><title>Shipping Label - {{.TrackingNumber}}</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:Arial,sans-serif;color:#333;line-height:1.4}
.shipping-label{max-width:800px;margin:0 auto;background:white;border:2px solid #e0e0e0;border-radius:8px;overflow:hidden}
.tracking-section{background:#f8f9fa;padding:15px 20px;border-bottom:1px solid #e0e0e0;display:flex;justify-content:space-between;align-items:center}
.tracking-info h2{font-size:24px;color:#2c3e50;margin-bottom:5px}
.tracking-info p{color:#666;font-size:14px}
.qr-code-section{text-align:center}
.qr-code-section img{width:80px;height:80px;border:2px solid #ddd;border-radius:4px;background:white}
.qr-label{font-size:10px;margin-top:5px;color:#666}
.main-content{padding:20px}
.addresses-section{display:grid;grid-template-columns:1fr 1fr;gap:30px;margin-bottom:30px}
.address-box{border:1px solid #ddd;border-radius:6px;padding:20px;background:#fafafa}
.address-box h3{color:#2c3e50;font-size:18px;margin-bottom:15px;padding-bottom:8px;border-bottom:2px solid #3498db;display:inline-block}
.address-detail{margin-bottom:8px;font-size:14px}
.address-detail strong{color:#2c3e50;display:inline-block;width:80px}
.package-details{background:#fff;border:1px solid #ddd;border-radius:6px;padding:20px;margin-bottom:20px}
.package-details h3{color:#2c3e50;font-size:18px;margin-bottom:15px;padding-bottom:8px;border-bottom:2px solid #e74c3c;display:inline-block}
.details-grid{display:grid;grid-template-columns:1fr 1fr;gap:15px}
.detail-item{display:flex;justify-content:space-between;padding:8px 0;border-bottom:1px solid #eee}
.detail-item:last-child{border-bottom:none}
.detail-label{font-weight:bold;color:#2c3e50}
.detail-value{color:#555}
.currency{color:#27ae60;font-weight:bold}
.footer{background:#2c3e50;color:white;text-align:center;padding:15px;font-size:12px}
.footer p{margin-bottom:5px}
@media print{.shipping-label{border:none;margin:0;max-width:none}}
</style>
</head>
<body>
<div class="shipping-label">
  <div class="tracking-section">
    <div class="tracking-info">
      <h2>{{.TrackingNumber}}</h2>
      <p>Shipment ID: #{{.ShipmentID}} | Created: {{.CreatedDate}}</p>
    </div>
    <div class="qr-code-section">
      {{if .QRCodeBase64}}<img src="data:image/png;base64,{{.QRCodeBase64}}" alt="QR Code"/>{{end}}
      <p class="qr-label">Scan for tracking</p>
    </div>
  </div>
  <div class="main-content">
    <div class="addresses-section">
      <div class="address-box">
        <h3>SENDER DETAILS</h3>
        <div class="address-detail"><strong>Name:</strong> {{.SenderName}}</div>
        <div class="address-detail"><strong>Email:</strong> {{.SenderEmail}}</div>
        <div class="address-detail"><strong>Phone:</strong> {{.SenderPhone}}</div>
        <div class="address-detail"><strong>Address:</strong><br/>{{.PickupAddress}}</div>
      </div>
      <div class="address-box">
        <h3>RECIPIENT DETAILS</h3>
        <div class="address-detail"><strong>Name:</strong> {{.RecipientName}}</div>
        <div class="address-detail"><strong>Phone:</strong> {{.RecipientPhone}}</div>
        <div class="address-detail"><strong>Address:</strong><br/>{{.DestinationAddress}}</div>
      </div>
    </div>
    <div class="package-details">
      <h3>PACKAGE INFORMATION</h3>
      <div class="details-grid">
        <div class="detail-item"><span class="detail-label">Package Type:</span><span class="detail-value">{{.PackageType}}</span></div>
        <div class="detail-item"><span class="detail-label">Delivery Type:</span><span class="detail-value">{{.DeliveryType}}</span></div>
        <div class="detail-item"><span class="detail-label">Weight:</span><span class="detail-value">{{.Weight}}g</span></div>
        <div class="detail-item"><span class="detail-label">Distance:</span><span class="detail-value">{{.Distance}} km</span></div>
        <div class="detail-item"><span class="detail-label">Base Price:</span><span class="detail-value currency">IDR {{.BasePrice}}</span></div>
        <div class="detail-item"><span class="detail-label">Weight Price:</span><span class="detail-value currency">IDR {{.WeightPrice}}</span></div>
        <div class="detail-item"><span class="detail-label">Distance Price:</span><span class="detail-value currency">IDR {{.DistancePrice}}</span></div>
        <div class="detail-item"><span class="detail-label"><strong>Total:</strong></span><span class="detail-value currency"><strong>IDR {{.Price}}</strong></span></div>
      </div>
    </div>
  </div>
  <div class="footer">
    <p><strong>Kirim Aja Logistics</strong> - Your Trusted Delivery Partner</p>
    <p>Customer Service: +62-21-1234-5678 | Email: support@kirimaja.com</p>
    <p>Generated on {{.GeneratedDate}} | This is a computer-generated document</p>
  </div>
</div>
</body></html>`
