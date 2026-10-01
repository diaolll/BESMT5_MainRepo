package utils

import (
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"siakaduts/models"
)

type JWTManager struct {
	secret  []byte
	issuer  string
	ttlMin  int
}

func NewJWTManager(secret, issuer string, ttlMin int) *JWTManager {
	return &JWTManager{secret: []byte(secret), issuer: issuer, ttlMin: ttlMin}
}

func (m *JWTManager) TTLSeconds() int { return m.ttlMin * 60 }

type claims struct {
	Email string `json:"email"`
	Role  string `json:"role"`
	jwt.RegisteredClaims
}

func (m *JWTManager) Generate(u models.User) (string, error) {
	now := time.Now()
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		Email: u.Email,
		Role:  u.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.Itoa(u.ID),
			Issuer:    m.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(m.ttlMin) * time.Minute)),
		},
	})
	return tok.SignedString(m.secret)
}

func (m *JWTManager) Parse(token string) (models.AuthUser, bool, error) {
	cl := &claims{}
	t, err := jwt.ParseWithClaims(token, cl, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrTokenSignatureInvalid
		}
		return m.secret, nil
	}, jwt.WithIssuer(m.issuer), jwt.WithExpirationRequired())
	if err != nil || !t.Valid {
		expired := err != nil && (err.Error() == jwt.ErrTokenExpired.Error() || contains(err.Error(), "expired"))
		return models.AuthUser{}, expired, err
	}
	id, err := strconv.Atoi(cl.Subject)
	if err != nil {
		return models.AuthUser{}, false, err
	}
	return models.AuthUser{UserID: id, Email: cl.Email, Role: cl.Role}, false, nil
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (func() bool {
		for i := 0; i+len(sub) <= len(s); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	})()
}
