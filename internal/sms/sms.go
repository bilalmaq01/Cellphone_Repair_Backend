// Package sms sends text messages and runs OTP verification. It defines a
// Client interface with two implementations: a real Twilio client and a
// no-cost fake (LogClient) used when Twilio is not configured.
package sms

import "context"

// Client sends SMS messages and runs phone-number OTP verification.
type Client interface {
	// SendSMS sends a text message to the given number.
	SendSMS(ctx context.Context, to, body string) error
	// StartOTP begins verification by sending a one-time code to the phone.
	StartOTP(ctx context.Context, phone string) error
	// CheckOTP reports whether the code is valid for the phone.
	CheckOTP(ctx context.Context, phone, code string) (bool, error)
}
