// Package config loads application configuration from the environment,
// optionally seeded from a local .env file.
package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Config holds the settings the server needs to run.
type Config struct {
	DatabaseURL string
	JWTSecret   string

	// Port the HTTP server listens on (App Runner sets PORT). Defaults to 8080.
	Port string
	// StaticDir is the directory of built frontend assets to serve. Empty = API only.
	StaticDir string

	// Twilio (optional). When all four are set, real SMS/OTP is used; otherwise
	// the app falls back to a no-cost fake client that logs instead of texting.
	TwilioAccountSID string
	TwilioAuthToken  string
	TwilioVerifySID  string
	TwilioFrom       string

	// Supabase Storage (optional; needed for signature uploads).
	SupabaseURL        string
	SupabaseServiceKey string
	SignatureBucket    string
}

// TwilioConfigured reports whether all Twilio settings are present.
func (c *Config) TwilioConfigured() bool {
	return c.TwilioAccountSID != "" && c.TwilioAuthToken != "" &&
		c.TwilioVerifySID != "" && c.TwilioFrom != ""
}

// Load reads .env (if present) into the process environment, then builds a
// Config from environment variables. Real environment variables always win
// over values in .env.
func Load() (*Config, error) {
	// Best-effort: a missing .env is fine (e.g. in production, real env vars are set).
	_ = loadDotEnv(".env")

	cfg := &Config{
		DatabaseURL:        os.Getenv("DATABASE_URL"),
		JWTSecret:          os.Getenv("JWT_SECRET"),
		Port:               envOr("PORT", "8080"),
		StaticDir:          envOr("STATIC_DIR", "web/dist"),
		TwilioAccountSID:   os.Getenv("TWILIO_ACCOUNT_SID"),
		TwilioAuthToken:    os.Getenv("TWILIO_AUTH_TOKEN"),
		TwilioVerifySID:    os.Getenv("TWILIO_VERIFY_SERVICE_SID"),
		TwilioFrom:         os.Getenv("TWILIO_FROM_NUMBER"),
		SupabaseURL:        os.Getenv("SUPABASE_URL"),
		SupabaseServiceKey: os.Getenv("SUPABASE_SERVICE_ROLE_KEY"),
		SignatureBucket:    os.Getenv("SUPABASE_SIGNATURE_BUCKET"),
	}
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is not set (add it to .env or the environment)")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is not set (add it to .env or the environment)")
	}
	return cfg, nil
}

// envOr returns the environment variable value, or def if unset/empty.
func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// loadDotEnv parses KEY=VALUE lines from path into the process environment.
// Lines that are blank or start with '#' are ignored. Existing environment
// variables are not overwritten.
func loadDotEnv(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.Trim(strings.TrimSpace(val), `"'`)
		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, val)
		}
	}
	return sc.Err()
}
