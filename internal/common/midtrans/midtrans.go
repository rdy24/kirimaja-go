package midtrans

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"

	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
)

type Client struct {
	serverKey string
	env       midtrans.EnvironmentType
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
	return &Client{serverKey, e}
}

func (c *Client) CreateSnap(orderID string, amount int64, email string) (*SnapResult, error) {
	client := snap.Client{}
	client.New(c.serverKey, c.env)

	req := &snap.Request{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  orderID,
			GrossAmt: amount,
		},
		CustomerDetail: &midtrans.CustomerDetails{Email: email},
		Expiry:         &snap.ExpiryDetails{Duration: 24, Unit: "hour"},
	}

	resp, err := client.CreateTransaction(req)
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
		return expected == signatureKey
	}
}
