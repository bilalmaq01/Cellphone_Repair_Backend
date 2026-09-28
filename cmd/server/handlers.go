package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"CellphoneRepairBackend/internal/auth"
	"CellphoneRepairBackend/internal/models"
	"CellphoneRepairBackend/internal/phone"
	"CellphoneRepairBackend/internal/respond"
	"CellphoneRepairBackend/internal/sms"
	"CellphoneRepairBackend/internal/storage"
	"CellphoneRepairBackend/internal/store"
)

// server holds shared dependencies for the HTTP handlers.
type server struct {
	store   *store.Store
	auth    *auth.Manager
	sms     sms.Client
	storage *storage.Client // nil if Supabase Storage is not configured
}

// IntakeRequest is the JSON body submitted at repair intake (by an employee).
type IntakeRequest struct {
	Phone  string `json:"phone"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Device string `json:"device"`
	Issue  string `json:"issue"`

	// Signed authorization (optional unless Storage is configured on the client).
	Signature    string `json:"signature"`      // base64 PNG, optionally a data URL
	Terms        string `json:"terms"`          // terms text shown at signing
	SignedByName string `json:"signed_by_name"` // defaults to Name
}

// health is a simple liveness check.
func (s *server) health(w http.ResponseWriter, r *http.Request) {
	respond.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// intake creates (or updates) the customer, creates a repair, and — if a
// signature was captured — uploads it and records the signed authorization.
func (s *server) intake(w http.ResponseWriter, r *http.Request) {
	var req IntakeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Name == "" || req.Device == "" || req.Issue == "" {
		respond.Error(w, http.StatusBadRequest, "name, device, and issue are required")
		return
	}

	normPhone, err := phone.Normalize(req.Phone)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid phone number")
		return
	}

	// If a signature was provided, upload it FIRST so a failure here doesn't
	// leave a dangling repair. objectPath doesn't depend on the repair id.
	var signaturePath string
	if req.Signature != "" {
		if s.storage == nil {
			respond.Error(w, http.StatusServiceUnavailable, "signature capture is not configured (Supabase Storage)")
			return
		}
		data, err := decodeSignature(req.Signature)
		if err != nil {
			respond.Error(w, http.StatusBadRequest, "invalid signature image")
			return
		}
		path := fmt.Sprintf("%s-%d.png", strings.TrimPrefix(normPhone, "+"), time.Now().UnixNano())
		if _, err := s.storage.Upload(r.Context(), path, data, "image/png"); err != nil {
			log.Printf("intake: signature upload failed: %v", err)
			respond.Error(w, http.StatusInternalServerError, "could not upload signature")
			return
		}
		signaturePath = path
	}

	var email *string
	if req.Email != "" {
		email = &req.Email
	}
	if _, err := s.store.UpsertCustomer(r.Context(), models.Customer{
		PhoneNumber: normPhone,
		Name:        req.Name,
		Email:       email,
	}); err != nil {
		respond.Error(w, http.StatusInternalServerError, "could not save customer")
		return
	}

	// Record which logged-in employee performed the intake.
	var intakeEmployeeID *int64
	if claims, ok := auth.ClaimsFromContext(r.Context()); ok {
		intakeEmployeeID = &claims.EmployeeID
	}

	repair, err := s.store.CreateRepair(r.Context(), models.Repair{
		CustomerPhone:    normPhone,
		Device:           req.Device,
		IssueDescription: req.Issue,
		IntakeEmployeeID: intakeEmployeeID,
	})
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "could not create repair")
		return
	}

	if signaturePath != "" {
		signedBy := req.SignedByName
		if signedBy == "" {
			signedBy = req.Name
		}
		terms := req.Terms
		if terms == "" {
			terms = "(terms wording pending)"
		}
		if _, err := s.store.CreateAuthorization(r.Context(), models.RepairAuthorization{
			RepairID:     repair.ID,
			TermsText:    terms,
			SignatureURL: signaturePath,
			SignedByName: signedBy,
			EmployeeID:   intakeEmployeeID,
		}); err != nil {
			respond.Error(w, http.StatusInternalServerError, "could not save signed authorization")
			return
		}
	}

	respond.JSON(w, http.StatusCreated, repair)
}

// decodeSignature accepts a raw base64 string or a data URL and returns the bytes.
func decodeSignature(s string) ([]byte, error) {
	if i := strings.Index(s, ","); strings.HasPrefix(s, "data:") && i >= 0 {
		s = s[i+1:]
	}
	return base64.StdEncoding.DecodeString(s)
}

// getAuthorization returns the signed authorization for a repair, including a
// short-lived signed URL to view the signature image.
func (s *server) getAuthorization(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid repair id")
		return
	}
	authz, err := s.store.GetAuthorizationByRepair(r.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		respond.Error(w, http.StatusNotFound, "no signed authorization for this repair")
		return
	}
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "could not load authorization")
		return
	}

	var signedURL string
	if s.storage != nil {
		if url, err := s.storage.SignedURL(r.Context(), authz.SignatureURL, 300); err == nil {
			signedURL = url
		} else {
			log.Printf("authorization: signed url failed: %v", err)
		}
	}

	respond.JSON(w, http.StatusOK, map[string]any{
		"repair_id":      authz.RepairID,
		"terms_text":     authz.TermsText,
		"signed_by_name": authz.SignedByName,
		"signed_at":      authz.SignedAt,
		"signature_url":  signedURL, // short-lived viewing URL (300s)
	})
}

// getRepair returns a single repair by id.
func (s *server) getRepair(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid repair id")
		return
	}
	repair, err := s.store.GetRepair(r.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		respond.Error(w, http.StatusNotFound, "repair not found")
		return
	}
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "could not load repair")
		return
	}
	respond.JSON(w, http.StatusOK, repair)
}

// statusUpdateRequest is the body for changing a repair's status.
type statusUpdateRequest struct {
	Status string `json:"status"`
}

// updateRepairStatus changes a repair's status.
func (s *server) updateRepairStatus(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid repair id")
		return
	}
	var req statusUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if !models.IsValidStatus(req.Status) {
		respond.Error(w, http.StatusBadRequest, "invalid status")
		return
	}

	repair, err := s.store.UpdateRepairStatus(r.Context(), id, req.Status)
	if errors.Is(err, store.ErrNotFound) {
		respond.Error(w, http.StatusNotFound, "repair not found")
		return
	}
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "could not update status")
		return
	}

	s.notifyStatusChange(r.Context(), repair)
	respond.JSON(w, http.StatusOK, repair)
}

// notifyStatusChange texts the customer about the new status and logs the
// notification. Best-effort: failures are recorded but do not fail the request.
func (s *server) notifyStatusChange(ctx context.Context, repair models.Repair) {
	human := strings.ReplaceAll(strings.ToLower(repair.Status), "_", " ")
	msg := fmt.Sprintf("Your %s repair is now %s.", repair.Device, human)

	status := "sent"
	if err := s.sms.SendSMS(ctx, repair.CustomerPhone, msg); err != nil {
		status = "failed"
		log.Printf("notify: sending sms failed: %v", err)
	}
	if err := s.store.CreateNotification(ctx, repair.CustomerPhone, repair.ID, msg, status); err != nil {
		log.Printf("notify: recording notification failed: %v", err)
	}
}

// listRepairs powers the dashboard. ?open=true returns only the open work queue.
func (s *server) listRepairs(w http.ResponseWriter, r *http.Request) {
	openOnly := r.URL.Query().Get("open") == "true"
	repairs, err := s.store.ListRepairs(r.Context(), openOnly)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "could not list repairs")
		return
	}
	respond.JSON(w, http.StatusOK, repairs)
}
