package midtrans

import (
	"crypto/sha512"
	"crypto/subtle"
	"encoding/hex"
	"fmt"

	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
)

type Client struct {
	serverKey string
	env       midtrans.EnvironmentType
	snap      snap.Client
}

type SnapResult struct {
	Token       string
	RedirectURL string
}

func New(serverKey, env string) *Client {
	e := midtrans.Sandbox
	if env == "production" {
		e = midtrans.Production
	}
	c := &Client{serverKey: serverKey, env: e}
	c.snap.New(serverKey, e) // build the HTTP client once, reuse per request
	return c
}

func (c *Client) CreateSnap(orderID string, amount int64, email string) (*SnapResult, error) {
	req := &snap.Request{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  orderID,
			GrossAmt: amount,
		},
		CustomerDetail: &midtrans.CustomerDetails{Email: email},
		Expiry:         &snap.ExpiryDetails{Duration: 24, Unit: "hour"},
	}

	resp, err := c.snap.CreateTransaction(req)
	if err != nil {
		return nil, fmt.Errorf("midtrans snap error: %w", err)
	}
	return &SnapResult{Token: resp.Token, RedirectURL: resp.RedirectURL}, nil
}

func (c *Client) VerifyWebhookSignature(orderID, statusCode, grossAmount string) func(signatureKey string) bool {
	raw := orderID + statusCode + grossAmount + c.serverKey
	hash := sha512.Sum512([]byte(raw))
	expected := hex.EncodeToString(hash[:])
	return func(signatureKey string) bool {
		return subtle.ConstantTimeCompare([]byte(expected), []byte(signatureKey)) == 1
	}
}
