// Command seedadmin creates the first admin account.
// Usage: go run ./cmd/seedadmin -email you@shop.com -password 'your-password'
package main

import (
	"context"
	"flag"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"

	"CellphoneRepairBackend/internal/auth"
	"CellphoneRepairBackend/internal/config"
	"CellphoneRepairBackend/internal/store"
)

func main() {
	email := flag.String("email", "", "admin email")
	password := flag.String("password", "", "admin password")
	flag.Parse()

	if *email == "" || *password == "" {
		log.Fatal("usage: go run ./cmd/seedadmin -email you@shop.com -password 'your-password'")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer pool.Close()

	hash, err := auth.HashPassword(*password)
	if err != nil {
		log.Fatalf("hash password: %v", err)
	}

	emp, err := store.New(pool).CreateEmployee(ctx, *email, hash, "admin")
	if err != nil {
		log.Fatalf("create admin (email may already exist): %v", err)
	}
	log.Printf("created admin id=%d email=%s", emp.ID, emp.Email)
}
