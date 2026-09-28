package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"CellphoneRepairBackend/internal/auth"
	"CellphoneRepairBackend/internal/config"
	"CellphoneRepairBackend/internal/db"
	"CellphoneRepairBackend/internal/sms"
	"CellphoneRepairBackend/internal/storage"
	"CellphoneRepairBackend/internal/store"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx := context.Background()
	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()
	log.Println("db connected")

	authMgr := auth.NewManager(cfg.JWTSecret, 12*time.Hour)

	var smsClient sms.Client
	if cfg.TwilioConfigured() {
		smsClient = sms.NewTwilioClient(cfg.TwilioAccountSID, cfg.TwilioAuthToken, cfg.TwilioVerifySID, cfg.TwilioFrom)
		log.Println("sms: using Twilio")
	} else {
		smsClient = sms.NewLogClient()
		log.Println("sms: using fake (log) client — set Twilio creds in .env to send real texts")
	}

	var storageClient *storage.Client
	if cfg.SupabaseURL != "" && cfg.SupabaseServiceKey != "" && cfg.SignatureBucket != "" {
		storageClient = storage.New(cfg.SupabaseURL, cfg.SupabaseServiceKey, cfg.SignatureBucket)
		log.Println("storage: Supabase Storage enabled for signatures")
	} else {
		log.Println("storage: not configured — signature capture disabled")
	}

	srv := &server{store: store.New(pool), auth: authMgr, sms: smsClient, storage: storageClient}

	mux := http.NewServeMux()

	// Public
	mux.HandleFunc("GET /health", srv.health)
	mux.HandleFunc("POST /api/auth/login", srv.login)
	mux.HandleFunc("POST /api/status/request-otp", srv.requestOTP)
	mux.HandleFunc("POST /api/status/verify-otp", srv.verifyOTP)

	// Authenticated (any logged-in employee)
	authed := func(h http.HandlerFunc) http.Handler {
		return authMgr.RequireAuth(h)
	}
	mux.Handle("GET /api/auth/me", authed(srv.me))
	mux.Handle("POST /api/intake", authed(srv.intake))
	mux.Handle("GET /api/repairs", authed(srv.listRepairs))
	mux.Handle("GET /api/repairs/{id}", authed(srv.getRepair))
	mux.Handle("PATCH /api/repairs/{id}/status", authed(srv.updateRepairStatus))
	mux.Handle("GET /api/repairs/{id}/authorization", authed(srv.getAuthorization))

	// Customer status check (short-lived customer token from OTP)
	mux.Handle("GET /api/status/repairs", authMgr.RequireCustomer(http.HandlerFunc(srv.customerRepairs)))

	// Admin only
	admin := func(h http.HandlerFunc) http.Handler {
		return authMgr.RequireAuth(auth.RequireAdmin(h))
	}
	mux.Handle("GET /api/employees", admin(srv.listEmployees))
	mux.Handle("POST /api/employees", admin(srv.createEmployee))
	mux.Handle("PATCH /api/employees/{id}", admin(srv.updateEmployee))

	// Serve the built React app (and SPA fallback) when the assets are present.
	if dirExists(cfg.StaticDir) {
		mux.Handle("/", spaHandler(cfg.StaticDir))
		log.Printf("serving frontend from %s", cfg.StaticDir)
	} else {
		log.Printf("no frontend assets at %s — running API only", cfg.StaticDir)
	}

	// Wrap the whole mux so every request is logged.
	handler := loggingMiddleware(mux)

	addr := ":" + cfg.Port
	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal(err)
	}
}
