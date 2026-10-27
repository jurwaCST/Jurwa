package domain

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSecret = "test-secret-at-least-32-bytes-long!!"

func newTestIssuer(t *testing.T, now time.Time) *TokenIssuer {
	t.Helper()
	ti, err := NewTokenIssuer([]byte(testSecret), 15*time.Minute)
	require.NoError(t, err)
	ti.now = func() time.Time { return now }
	return ti
}

func TestNewTokenIssuerRejectsWeakConfig(t *testing.T) {
	tests := []struct {
		name   string
		secret string
		ttl    time.Duration
	}{
		{name: "empty secret", secret: "", ttl: time.Minute},
		{name: "secret shorter than 32 bytes", secret: "too-short", ttl: time.Minute},
		{name: "zero expiry", secret: testSecret, ttl: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewTokenIssuer([]byte(tt.secret), tt.ttl)
			assert.Error(t, err)
		})
	}
}

func TestIssueAndParse(t *testing.T) {
	ti := newTestIssuer(t, fixedNow)
	u := User{ID: "user-1", Email: "gayley@rub.edu.bt", Roles: []Role{RoleStudent}, Active: true}

	token, expiresAt, err := ti.Issue(u)
	require.NoError(t, err)
	assert.Equal(t, fixedNow.Add(15*time.Minute), expiresAt)

	claims, err := ti.Parse(token)
	require.NoError(t, err)
	assert.Equal(t, "user-1", claims.UserID)
	assert.Equal(t, "gayley@rub.edu.bt", claims.Email)
	assert.Equal(t, []Role{RoleStudent}, claims.Roles)
}

func TestParseRejectsBadTokens(t *testing.T) {
	issuedAt := fixedNow
	ti := newTestIssuer(t, issuedAt)
	u := User{ID: "user-1", Email: "gayley@rub.edu.bt", Roles: []Role{RoleStudent}, Active: true}
	token, _, err := ti.Issue(u)
	require.NoError(t, err)

	other, err := NewTokenIssuer([]byte("another-secret-that-is-32-bytes-long"), 15*time.Minute)
	require.NoError(t, err)
	other.now = ti.now
	otherToken, _, err := other.Issue(u)
	require.NoError(t, err)

	parts := strings.Split(token, ".")
	require.Len(t, parts, 3)
	// Replace the payload with one claiming a different user, keeping the old signature.
	tamperedPayload := parts[0] + "." + "eyJzdWIiOiJhZG1pbiJ9" + "." + parts[2]
	// Change the last character of the signature.
	sig := []byte(parts[2])
	if sig[len(sig)-1] == 'A' {
		sig[len(sig)-1] = 'B'
	} else {
		sig[len(sig)-1] = 'A'
	}
	tamperedSignature := parts[0] + "." + parts[1] + "." + string(sig)
	// A token that says "alg: none" with no signature.
	noneToken := "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0." + parts[1] + "."

	tests := []struct {
		name  string
		token string
		at    time.Time
	}{
		{name: "expired token", token: token, at: issuedAt.Add(16 * time.Minute)},
		{name: "tampered payload", token: tamperedPayload, at: issuedAt},
		{name: "tampered signature", token: tamperedSignature, at: issuedAt},
		{name: "signed with another secret", token: otherToken, at: issuedAt},
		{name: "alg none", token: noneToken, at: issuedAt},
		{name: "empty", token: "", at: issuedAt},
		{name: "garbage", token: "not.a.jwt", at: issuedAt},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			at := tt.at
			ti.now = func() time.Time { return at }
			_, err := ti.Parse(tt.token)
			assert.ErrorIs(t, err, ErrInvalidToken)
		})
	}
}
