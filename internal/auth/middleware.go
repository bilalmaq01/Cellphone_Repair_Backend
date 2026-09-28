package auth

import (
	"context"
	"net/http"
	"strings"

	"CellphoneRepairBackend/internal/respond"
)

type ctxKey int

const (
	claimsKey ctxKey = iota
	customerClaimsKey
)

// RequireAuth verifies the Bearer token and stores the claims in the request
// context. Requests without a valid token get 401.
func (m *Manager) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		token, ok := strings.CutPrefix(header, "Bearer ")
		if !ok || token == "" {
			respond.Error(w, http.StatusUnauthorized, "missing or malformed Authorization header")
			return
		}
		claims, err := m.Parse(token)
		if err != nil {
			respond.Error(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}
		ctx := context.WithValue(r.Context(), claimsKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireAdmin must be wrapped inside RequireAuth. It rejects non-admin roles with 403.
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := ClaimsFromContext(r.Context())
		if !ok || claims.Role != "admin" {
			respond.Error(w, http.StatusForbidden, "admin access required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ClaimsFromContext returns the authenticated employee's claims, if present.
func ClaimsFromContext(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value(claimsKey).(*Claims)
	return claims, ok
}

// RequireCustomer verifies a customer status-check token and stores the phone
// in the request context.
func (m *Manager) RequireCustomer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		token, ok := strings.CutPrefix(header, "Bearer ")
		if !ok || token == "" {
			respond.Error(w, http.StatusUnauthorized, "missing or malformed Authorization header")
			return
		}
		claims, err := m.ParseCustomer(token)
		if err != nil {
			respond.Error(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}
		ctx := context.WithValue(r.Context(), customerClaimsKey, claims.Phone)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// CustomerPhoneFromContext returns the verified customer's phone, if present.
func CustomerPhoneFromContext(ctx context.Context) (string, bool) {
	phone, ok := ctx.Value(customerClaimsKey).(string)
	return phone, ok
}
