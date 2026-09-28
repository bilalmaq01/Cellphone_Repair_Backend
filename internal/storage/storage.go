// Package storage uploads files to Supabase Storage and creates short-lived
// signed URLs for viewing private objects.
package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client talks to the Supabase Storage REST API using the service-role key.
type Client struct {
	baseURL    string // e.g. https://xxxx.supabase.co
	serviceKey string
	bucket     string
	http       *http.Client
}

// New returns a Storage client for the given project URL, service-role key,
// and bucket name. The URL is normalized to the project base (any trailing
// "/rest/v1" or slash is stripped) so both the base URL and the data-API URL work.
func New(supabaseURL, serviceKey, bucket string) *Client {
	base := strings.TrimRight(supabaseURL, "/")
	base = strings.TrimSuffix(base, "/rest/v1")
	base = strings.TrimRight(base, "/")
	return &Client{
		baseURL:    base,
		serviceKey: serviceKey,
		bucket:     bucket,
		http:       &http.Client{Timeout: 20 * time.Second},
	}
}

// setAuth applies the headers Supabase's gateway requires: an apikey header
// plus a bearer token, both the service-role key.
func (c *Client) setAuth(req *http.Request) {
	req.Header.Set("apikey", c.serviceKey)
	req.Header.Set("Authorization", "Bearer "+c.serviceKey)
}

// Upload stores data at the given object path (e.g. "repair-42-169....png") and
// returns that path. contentType is the MIME type (e.g. "image/png").
func (c *Client) Upload(ctx context.Context, objectPath string, data []byte, contentType string) (string, error) {
	url := fmt.Sprintf("%s/storage/v1/object/%s/%s", c.baseURL, c.bucket, objectPath)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	c.setAuth(req)
	req.Header.Set("Content-Type", contentType)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("storage upload: status %d: %s", resp.StatusCode, string(body))
	}
	return objectPath, nil
}

// Delete removes an object from the configured bucket.
func (c *Client) Delete(ctx context.Context, objectPath string) error {
	url := fmt.Sprintf("%s/storage/v1/object/%s/%s", c.baseURL, c.bucket, objectPath)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	c.setAuth(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("storage delete: status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// SignedURL returns a temporary URL to view a private object, valid for the
// given number of seconds.
func (c *Client) SignedURL(ctx context.Context, objectPath string, expiresIn int) (string, error) {
	url := fmt.Sprintf("%s/storage/v1/object/sign/%s/%s", c.baseURL, c.bucket, objectPath)
	body, _ := json.Marshal(map[string]int{"expiresIn": expiresIn})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	c.setAuth(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("storage sign: status %d: %s", resp.StatusCode, string(respBody))
	}

	var out struct {
		SignedURL string `json:"signedURL"`
	}
	if err := json.Unmarshal(respBody, &out); err != nil {
		return "", err
	}
	// The API returns a relative path; make it absolute.
	return c.baseURL + "/storage/v1" + out.SignedURL, nil
}
