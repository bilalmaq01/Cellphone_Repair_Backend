// Package models holds the core data types shared across the app.
package models

import "time"

// repairStatuses is the set of valid repair status values (mirrors the
// repair_status enum in the database).
var repairStatuses = map[string]bool{
	"RECEIVED":          true,
	"DIAGNOSING":        true,
	"AWAITING_APPROVAL": true,
	"IN_PROGRESS":       true,
	"READY_FOR_PICKUP":  true,
	"COMPLETED":         true,
	"CANCELLED":         true,
}

// IsValidStatus reports whether s is a recognized repair status.
func IsValidStatus(s string) bool {
	return repairStatuses[s]
}

// RepairAuthorization is the signed agreement captured at intake.
type RepairAuthorization struct {
	ID           int64     `json:"id"`
	RepairID     int64     `json:"repair_id"`
	TermsText    string    `json:"terms_text"`
	SignatureURL string    `json:"signature_url"` // object path in Supabase Storage
	SignedByName string    `json:"signed_by_name"`
	EmployeeID   *int64    `json:"employee_id,omitempty"`
	SignedAt     time.Time `json:"signed_at"`
}

// Employee is a staff account that can log in. Role is "employee" or "admin".
type Employee struct {
	ID           int64     `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // never serialized to JSON
	Role         string    `json:"role"`
	Active       bool      `json:"active"`
	CreatedAt    time.Time `json:"created_at"`
}

// Customer is a person we do repairs for, uniquely identified by phone number.
type Customer struct {
	PhoneNumber string    `json:"phone_number"`
	Name        string    `json:"name"`
	Email       *string   `json:"email,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Repair is a single repair job for a customer's device.
type Repair struct {
	ID                  int64      `json:"id"`
	CustomerPhone       string     `json:"customer_phone"`
	Device              string     `json:"device"`
	IssueDescription    string     `json:"issue_description"`
	Status              string     `json:"status"`
	Price               *float64   `json:"price,omitempty"`
	Notes               *string    `json:"notes,omitempty"`
	PartsUsed           *string    `json:"parts_used,omitempty"`
	Warranty            *string    `json:"warranty,omitempty"`
	EstimatedCompletion *time.Time `json:"estimated_completion,omitempty"`
	IntakeEmployeeID    *int64     `json:"intake_employee_id,omitempty"`
	FrontPhotoPath      *string    `json:"-"`
	BackPhotoPath       *string    `json:"-"`
	FrontPhotoURL       string     `json:"front_photo_url,omitempty"`
	BackPhotoURL        string     `json:"back_photo_url,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}
