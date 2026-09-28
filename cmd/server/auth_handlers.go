package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"CellphoneRepairBackend/internal/auth"
	"CellphoneRepairBackend/internal/respond"
	"CellphoneRepairBackend/internal/store"
)

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// login verifies credentials and returns a JWT.
func (s *server) login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	emp, err := s.store.GetEmployeeByEmail(r.Context(), req.Email)
	// Use one generic message so we don't reveal whether the email exists.
	if errors.Is(err, store.ErrNotFound) || (err == nil && !auth.CheckPassword(emp.PasswordHash, req.Password)) {
		respond.Error(w, http.StatusUnauthorized, "invalid email or password")
		return
	}
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "login failed")
		return
	}
	if !emp.Active {
		respond.Error(w, http.StatusForbidden, "account is deactivated")
		return
	}

	token, err := s.auth.Generate(emp)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "could not issue token")
		return
	}
	respond.JSON(w, http.StatusOK, map[string]string{"token": token})
}

// me returns the currently authenticated employee.
func (s *server) me(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "not authenticated")
		return
	}
	emp, err := s.store.GetEmployeeByID(r.Context(), claims.EmployeeID)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "could not load account")
		return
	}
	respond.JSON(w, http.StatusOK, emp)
}

type createEmployeeRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

// createEmployee (admin) creates a new staff account.
func (s *server) createEmployee(w http.ResponseWriter, r *http.Request) {
	var req createEmployeeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Email == "" || req.Password == "" {
		respond.Error(w, http.StatusBadRequest, "email and password are required")
		return
	}
	role := req.Role
	if role == "" {
		role = "employee"
	}
	if role != "employee" && role != "admin" {
		respond.Error(w, http.StatusBadRequest, "role must be 'employee' or 'admin'")
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "could not hash password")
		return
	}
	emp, err := s.store.CreateEmployee(r.Context(), req.Email, hash, role)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "could not create employee (email may already exist)")
		return
	}
	respond.JSON(w, http.StatusCreated, emp)
}

// listEmployees (admin) returns all staff accounts.
func (s *server) listEmployees(w http.ResponseWriter, r *http.Request) {
	employees, err := s.store.ListEmployees(r.Context())
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "could not list employees")
		return
	}
	respond.JSON(w, http.StatusOK, employees)
}

type updateEmployeeRequest struct {
	Role     *string `json:"role"`
	Active   *bool   `json:"active"`
	Password *string `json:"password"`
}

// updateEmployee (admin) changes role, active status, and/or resets the password.
func (s *server) updateEmployee(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid employee id")
		return
	}
	var req updateEmployeeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Role != nil && *req.Role != "employee" && *req.Role != "admin" {
		respond.Error(w, http.StatusBadRequest, "role must be 'employee' or 'admin'")
		return
	}

	var passwordHash *string
	if req.Password != nil {
		hash, err := auth.HashPassword(*req.Password)
		if err != nil {
			respond.Error(w, http.StatusInternalServerError, "could not hash password")
			return
		}
		passwordHash = &hash
	}

	emp, err := s.store.UpdateEmployee(r.Context(), id, req.Role, req.Active, passwordHash)
	if errors.Is(err, store.ErrNotFound) {
		respond.Error(w, http.StatusNotFound, "employee not found")
		return
	}
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "could not update employee")
		return
	}
	respond.JSON(w, http.StatusOK, emp)
}
