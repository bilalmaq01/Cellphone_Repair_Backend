// Package auth handles password hashing and JWT issuing/verification for
// employee accounts.
package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"CellphoneRepairBackend/internal/models"
)

// HashPassword returns a bcrypt hash of the given plaintext password.
func HashPassword(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// CheckPassword reports whether password matches the stored bcrypt hash.
func CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// Claims are the JWT claims we embed for an authenticated employee.
type Claims struct {
	EmployeeID int64  `json:"eid"`
	Role       string `json:"role"`
	jwt.RegisteredClaims
}

// Manager issues and verifies JWTs signed with a shared secret.
type Manager struct {
	secret []byte
	ttl    time.Duration
}

// NewManager returns a Manager. Tokens are valid for the given ttl.
func NewManager(secret string, ttl time.Duration) *Manager {
	return &Manager{secret: []byte(secret), ttl: ttl}
}

// Generate creates a signed JWT for the given employee.
func (m *Manager) Generate(emp models.Employee) (string, error) {
	now := time.Now()
	claims := Claims{
		EmployeeID: emp.ID,
		Role:       emp.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d", emp.ID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

// Parse verifies a token string and returns its claims.
func (m *Manager) Parse(tokenString string) (*Claims, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, m.keyFunc)
	if err != nil {
		return nil, err
	}
	return claims, nil
}

func (m *Manager) keyFunc(t *jwt.Token) (any, error) {
	if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
	}
	return m.secret, nil
}

// CustomerClaims are the claims for a short-lived customer status-check token.
type CustomerClaims struct {
	Phone string `json:"phone"`
	jwt.RegisteredClaims
}

// customerTokenTTL is how long a customer status-check token stays valid.
const customerTokenTTL = 15 * time.Minute

// GenerateCustomer issues a short-lived token letting a verified customer view
// their repairs by phone number.
func (m *Manager) GenerateCustomer(phone string) (string, error) {
	now := time.Now()
	claims := CustomerClaims{
		Phone: phone,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "customer:" + phone,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(customerTokenTTL)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

// ParseCustomer verifies a customer token and returns its claims.
func (m *Manager) ParseCustomer(tokenString string) (*CustomerClaims, error) {
	claims := &CustomerClaims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, m.keyFunc)
	if err != nil {
		return nil, err
	}
	return claims, nil
}
