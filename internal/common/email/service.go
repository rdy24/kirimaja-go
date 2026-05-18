package email

import (
	"bytes"
	"fmt"
	"html/template"
	"strconv"
	"time"

	"gopkg.in/gomail.v2"
)

type Service struct {
	host   string
	port   int
	user   string
	pass   string
	sender string
}

func New(host string, port int, user, pass, sender string) *Service {
	if sender == "" {
		sender = user
	}
	return &Service{host, port, user, pass, sender}
}

func (s *Service) send(to, subject, html string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", s.sender)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", html)
	d := gomail.NewDialer(s.host, s.port, s.user, s.pass)
	return d.DialAndSend(m)
}

func formatIDR(amount float64) string {
	n := int64(amount)
	s := strconv.FormatInt(n, 10)
	out := make([]byte, 0, len(s)+len(s)/3)
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, '.')
		}
		out = append(out, byte(c))
	}
	return string(out)
}

type notifData struct {
	ShipmentID uint
	PaymentURL string
	Amount     string
	ExpiryDate string
}

type successData struct {
	ShipmentID     uint
	Amount         string
	PaymentDate    string
	TrackingNumber string
}

var notifTmpl = template.Must(template.New("notif").Parse(`<html lang="en"><head><meta charset="UTF-8"/><title>Payment Required</title></head>
<body><div style="font-family:Arial,sans-serif;max-width:600px;margin:0 auto;padding:20px">
<div style="text-align:center;margin-bottom:30px"><h1 style="color:#333;margin:0">Kirim Aja</h1>
<p style="color:#666;margin:5px 0">Reliable Shipping Service</p></div>
<h2 style="color:#333">Payment Required</h2><p>Dear Customer,</p>
<p>Thank you for using our shipping service. A payment is required for your shipment.</p>
<div style="background-color:#f5f5f5;padding:20px;border-radius:8px;margin:20px 0;border-left:4px solid #007bff">
<h3 style="margin:0 0 15px 0;color:#333">Shipment Details:</h3>
<table style="width:100%;border-collapse:collapse">
<tr><td style="padding:5px 0;color:#666"><strong>Shipment ID:</strong></td><td style="padding:5px 0;color:#333">#{{.ShipmentID}}</td></tr>
<tr><td style="padding:5px 0;color:#666"><strong>Amount:</strong></td><td style="padding:5px 0;color:#333">Rp {{.Amount}}</td></tr>
<tr><td style="padding:5px 0;color:#666"><strong>Payment Due Date:</strong></td><td style="padding:5px 0;color:#333">{{.ExpiryDate}}</td></tr>
</table></div>
<p>Please click the button below to complete your payment:</p>
<div style="text-align:center;margin:30px 0">
<a href="{{.PaymentURL}}" style="background-color:#007bff;color:white;padding:15px 30px;text-decoration:none;border-radius:8px;display:inline-block;font-weight:bold">Pay Now</a>
</div>
<hr style="border:none;border-top:1px solid #eee;margin:30px 0"/>
<p style="color:#666">If you have any questions, please contact our customer support.</p>
<div style="margin-top:30px;padding-top:20px;border-top:1px solid #eee">
<p style="margin:0;color:#333">Best regards,</p>
<p style="margin:5px 0 0 0;color:#007bff;font-weight:bold">Kirim Aja Team</p></div>
</div></body></html>`))

var successTmpl = template.Must(template.New("success").Parse(`<html lang="en"><head><meta charset="UTF-8"/><title>Payment Successful</title></head>
<body><div style="font-family:Arial,sans-serif;max-width:600px;margin:0 auto;padding:20px">
<div style="text-align:center;margin-bottom:30px"><h1 style="color:#333;margin:0">Kirim Aja</h1>
<p style="color:#666;margin:5px 0">Reliable Shipping Service</p></div>
<div style="text-align:center;margin-bottom:20px">
<h2 style="color:#28a745;margin:0">Payment Successful!</h2></div>
<p>Dear Customer,</p><p>Great news! Your payment has been successfully processed.</p>
<div style="background-color:#d4edda;padding:20px;border-radius:8px;margin:20px 0;border-left:4px solid #28a745">
<h3 style="margin:0 0 15px 0;color:#155724">Payment Details:</h3>
<table style="width:100%;border-collapse:collapse">
<tr><td style="padding:5px 0;color:#155724"><strong>Shipment ID:</strong></td><td style="padding:5px 0;color:#155724">#{{.ShipmentID}}</td></tr>
<tr><td style="padding:5px 0;color:#155724"><strong>Amount Paid:</strong></td><td style="padding:5px 0;color:#155724">Rp {{.Amount}}</td></tr>
<tr><td style="padding:5px 0;color:#155724"><strong>Tracking Number:</strong></td><td style="padding:5px 0;color:#155724;font-weight:bold">{{.TrackingNumber}}</td></tr>
<tr><td style="padding:5px 0;color:#155724"><strong>Payment Date:</strong></td><td style="padding:5px 0;color:#155724">{{.PaymentDate}}</td></tr>
</table></div>
<hr style="border:none;border-top:1px solid #eee;margin:30px 0"/>
<p style="color:#666">Thank you for choosing our service!</p>
<div style="margin-top:30px;padding-top:20px;border-top:1px solid #eee">
<p style="margin:0;color:#333">Best regards,</p>
<p style="margin:5px 0 0 0;color:#007bff;font-weight:bold">Kirim Aja Team</p></div>
</div></body></html>`))

func (s *Service) SendPaymentNotification(to, paymentURL string, shipmentID uint, amount float64, expiryDate time.Time) error {
	var buf bytes.Buffer
	if err := notifTmpl.Execute(&buf, notifData{
		ShipmentID: shipmentID,
		PaymentURL: paymentURL,
		Amount:     formatIDR(amount),
		ExpiryDate: expiryDate.Format("02/01/2006"),
	}); err != nil {
		return err
	}
	return s.send(to, fmt.Sprintf("Payment Notification for Shipment #%d", shipmentID), buf.String())
}

func (s *Service) SendPaymentSuccess(to string, shipmentID uint, amount float64, trackingNumber string) error {
	var buf bytes.Buffer
	if err := successTmpl.Execute(&buf, successData{
		ShipmentID:     shipmentID,
		Amount:         formatIDR(amount),
		PaymentDate:    time.Now().Format("02/01/2006"),
		TrackingNumber: trackingNumber,
	}); err != nil {
		return err
	}
	return s.send(to, fmt.Sprintf("Payment Successful - Shipment #%d", shipmentID), buf.String())
}
