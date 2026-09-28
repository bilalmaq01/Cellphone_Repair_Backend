package sms

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"sync"
)

// LogClient is a no-cost fake used in development when Twilio is not configured.
// It logs instead of texting. OTP codes are generated, logged, and held in
// memory so the verification flow works end-to-end without Twilio.
type LogClient struct {
	mu    sync.Mutex
	codes map[string]string
}

// NewLogClient returns a ready-to-use fake SMS client.
func NewLogClient() *LogClient {
	return &LogClient{codes: make(map[string]string)}
}

func (c *LogClient) SendSMS(ctx context.Context, to, body string) error {
	log.Printf("[fake-sms] to=%s body=%q", to, body)
	return nil
}

func (c *LogClient) StartOTP(ctx context.Context, phone string) error {
	code := randomCode()
	c.mu.Lock()
	c.codes[phone] = code
	c.mu.Unlock()
	log.Printf("[fake-sms] OTP for %s is %s (dev only — real Twilio would text this)", phone, code)
	return nil
}

func (c *LogClient) CheckOTP(ctx context.Context, phone, code string) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if want, ok := c.codes[phone]; ok && want == code {
		delete(c.codes, phone)
		return true, nil
	}
	return false, nil
}

func randomCode() string {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "000000"
	}
	return fmt.Sprintf("%06d", n.Int64())
}
