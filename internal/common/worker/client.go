package worker

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/hibiken/asynq"
)

type Client struct {
	c         *asynq.Client
	inspector *asynq.Inspector
}

func NewClient(redisAddr string) *Client {
	opt := asynq.RedisClientOpt{Addr: redisAddr}
	return &Client{
		c:         asynq.NewClient(opt),
		inspector: asynq.NewInspector(opt),
	}
}

func (c *Client) enqueue(typeName string, payload any, opts ...asynq.Option) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = c.c.Enqueue(asynq.NewTask(typeName, b), opts...)
	return err
}

func (c *Client) EnqueuePaymentNotification(p PaymentNotificationPayload) error {
	return c.enqueue(TypePaymentNotification, p, asynq.MaxRetry(3))
}

func (c *Client) EnqueuePaymentSuccess(p PaymentSuccessPayload) error {
	return c.enqueue(TypePaymentSuccess, p, asynq.MaxRetry(3))
}

func (c *Client) EnqueuePaymentExpiry(p PaymentExpiryPayload, at time.Time) error {
	delay := time.Until(at)
	if delay < 0 {
		delay = 0
	}
	taskID := fmt.Sprintf("payment:expiry:%d", p.PaymentID)
	return c.enqueue(TypePaymentExpiry, p,
		asynq.ProcessIn(delay),
		asynq.TaskID(taskID),
		asynq.MaxRetry(3),
	)
}

func (c *Client) CancelPaymentExpiry(paymentID uint) {
	taskID := fmt.Sprintf("payment:expiry:%d", paymentID)
	if err := c.inspector.DeleteTask(DefaultQueue, taskID); err != nil {
		log.Printf("[worker] cancel payment expiry task %s: %v", taskID, err)
	}
}
