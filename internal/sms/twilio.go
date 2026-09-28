package sms

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// TwilioClient talks to the Twilio Messages and Verify REST APIs. It is used
// when Twilio credentials are configured; otherwise the app uses LogClient.
type TwilioClient struct {
	accountSID string
	authToken  string
	verifySID  string
	from       string
	http       *http.Client
}

// NewTwilioClient builds a Twilio-backed SMS client.
func NewTwilioClient(accountSID, authToken, verifySID, from string) *TwilioClient {
	return &TwilioClient{
		accountSID: accountSID,
		authToken:  authToken,
		verifySID:  verifySID,
		from:       from,
		http:       http.DefaultClient,
	}
}

func (c *TwilioClient) SendSMS(ctx context.Context, to, body string) error {
	endpoint := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", c.accountSID)
	form := url.Values{}
	form.Set("To", to)
	form.Set("From", c.from)
	form.Set("Body", body)
	return c.post(ctx, endpoint, form, nil)
}

func (c *TwilioClient) StartOTP(ctx context.Context, phone string) error {
	endpoint := fmt.Sprintf("https://verify.twilio.com/v2/Services/%s/Verifications", c.verifySID)
	form := url.Values{}
	form.Set("To", phone)
	form.Set("Channel", "sms")
	return c.post(ctx, endpoint, form, nil)
}

func (c *TwilioClient) CheckOTP(ctx context.Context, phone, code string) (bool, error) {
	endpoint := fmt.Sprintf("https://verify.twilio.com/v2/Services/%s/VerificationCheck", c.verifySID)
	form := url.Values{}
	form.Set("To", phone)
	form.Set("Code", code)

	var result struct {
		Status string `json:"status"`
	}
	if err := c.post(ctx, endpoint, form, &result); err != nil {
		return false, err
	}
	return result.Status == "approved", nil
}

// post sends a form-encoded POST with basic auth and optionally decodes the
// JSON response into out.
func (c *TwilioClient) post(ctx context.Context, endpoint string, form url.Values, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.SetBasicAuth(c.accountSID, c.authToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("twilio: status %d: %s", resp.StatusCode, string(respBody))
	}
	if out != nil {
		return json.Unmarshal(respBody, out)
	}
	return nil
}
