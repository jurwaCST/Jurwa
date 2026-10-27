package domain

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const tokenIssuer = "jurwa-identity"

// MinSecretBytes is the shortest JWT_SECRET accepted for HS256.
const MinSecretBytes = 32

// Claims is what a valid token says about its holder.
type Claims struct {
	UserID string
	Email  string
	Roles  []Role
}

// TokenIssuer signs and checks JWTs with HS256 and a shared secret.
type TokenIssuer struct {
	secret []byte
	ttl    time.Duration
	now    func() time.Time
}

// NewTokenIssuer needs a secret of at least 32 bytes and a positive expiry.
func NewTokenIssuer(secret []byte, ttl time.Duration) (*TokenIssuer, error) {
	if len(secret) < MinSecretBytes {
		return nil, errors.New("JWT secret must be at least 32 bytes")
	}
	if ttl <= 0 {
		return nil, errors.New("token expiry must be positive")
	}
	return &TokenIssuer{secret: secret, ttl: ttl, now: time.Now}, nil
}

// SetClock replaces the clock. Used by tests to check expiry.
func (t *TokenIssuer) SetClock(now func() time.Time) { t.now = now }

type jwtClaims struct {
	Email string   `json:"email"`
	Roles []string `json:"roles"`
	jwt.RegisteredClaims
}

// Issue returns a signed token for the user and the time it expires.
func (t *TokenIssuer) Issue(u User) (string, time.Time, error) {
	now := t.now()
	expiresAt := now.Add(t.ttl)
	roles := make([]string, len(u.Roles))
	for i, r := range u.Roles {
		roles[i] = string(r)
	}
	claims := jwtClaims{
		Email: u.Email,
		Roles: roles,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    tokenIssuer,
			Subject:   u.ID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(t.secret)
	if err != nil {
		return "", time.Time{}, err
	}
	return signed, expiresAt, nil
}

// Parse checks the signature, algorithm, issuer and expiry. Any problem
// returns ErrInvalidToken.
func (t *TokenIssuer) Parse(token string) (Claims, error) {
	var c jwtClaims
	_, err := jwt.ParseWithClaims(token, &c,
		func(*jwt.Token) (any, error) { return t.secret, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(tokenIssuer),
		jwt.WithExpirationRequired(),
		jwt.WithTimeFunc(t.now),
	)
	if err != nil || c.Subject == "" {
		return Claims{}, ErrInvalidToken
	}
	roles := make([]Role, len(c.Roles))
	for i, r := range c.Roles {
		roles[i] = Role(r)
	}
	return Claims{UserID: c.Subject, Email: c.Email, Roles: roles}, nil
}
