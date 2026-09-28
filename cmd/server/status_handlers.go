package main

import (
	"encoding/json"
	"net/http"
	"time"

	"CellphoneRepairBackend/internal/auth"
	"CellphoneRepairBackend/internal/models"
	"CellphoneRepairBackend/internal/phone"
	"CellphoneRepairBackend/internal/respond"
)

// customerRepairView is the customer-safe subset of a repair (no internal notes).
type customerRepairView struct {
	ID                  int64      `json:"id"`
	Device              string     `json:"device"`
	Issue               string     `json:"issue"`
	Status              string     `json:"status"`
	Price               *float64   `json:"price,omitempty"`
	EstimatedCompletion *time.Time `json:"estimated_completion,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

func toCustomerView(r models.Repair) customerRepairView {
	return customerRepairView{
		ID:                  r.ID,
		Device:              r.Device,
		Issue:               r.IssueDescription,
		Status:              r.Status,
		Price:               r.Price,
		EstimatedCompletion: r.EstimatedCompletion,
		CreatedAt:           r.CreatedAt,
		UpdatedAt:           r.UpdatedAt,
	}
}

// requestOTP sends a one-time code to the customer's phone (public).
func (s *server) requestOTP(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Phone string `json:"phone"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	normPhone, err := phone.Normalize(req.Phone)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid phone number")
		return
	}
	// Start verification regardless of whether the number exists, so we don't
	// leak which numbers are customers.
	if err := s.sms.StartOTP(r.Context(), normPhone); err != nil {
		respond.Error(w, http.StatusInternalServerError, "could not send code")
		return
	}
	respond.JSON(w, http.StatusOK, map[string]string{"status": "otp_sent"})
}

// verifyOTP checks the code and, on success, returns a short-lived customer
// token plus the customer's repairs (public).
func (s *server) verifyOTP(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	normPhone, err := phone.Normalize(req.Phone)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid phone number")
		return
	}

	ok, err := s.sms.CheckOTP(r.Context(), normPhone, req.Code)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "could not verify code")
		return
	}
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "invalid or expired code")
		return
	}

	token, err := s.auth.GenerateCustomer(normPhone)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "could not issue token")
		return
	}

	repairs, err := s.store.ListRepairsByPhone(r.Context(), normPhone)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "could not load repairs")
		return
	}
	respond.JSON(w, http.StatusOK, map[string]any{
		"token":   token,
		"repairs": viewList(repairs),
	})
}

// customerRepairs returns the verified customer's repairs (customer token required).
func (s *server) customerRepairs(w http.ResponseWriter, r *http.Request) {
	phoneNum, ok := auth.CustomerPhoneFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "not authenticated")
		return
	}
	repairs, err := s.store.ListRepairsByPhone(r.Context(), phoneNum)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "could not load repairs")
		return
	}
	respond.JSON(w, http.StatusOK, viewList(repairs))
}

func viewList(repairs []models.Repair) []customerRepairView {
	views := make([]customerRepairView, 0, len(repairs))
	for _, r := range repairs {
		views = append(views, toCustomerView(r))
	}
	return views
}
