package domain

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var fixedNow = time.Date(2026, 10, 27, 9, 0, 0, 0, time.UTC)

func TestNormalizeEmail(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr error
	}{
		{name: "valid institutional email", input: "02240340.cst@rub.edu.bt", want: "02240340.cst@rub.edu.bt"},
		{name: "trims spaces and lowercases", input: "  Gayley.CST@RUB.edu.bt ", want: "gayley.cst@rub.edu.bt"},
		{name: "empty", input: "", wantErr: ErrInvalidEmail},
		{name: "no at sign", input: "gayley.rub.edu.bt", wantErr: ErrInvalidEmail},
		{name: "no domain", input: "gayley@", wantErr: ErrInvalidEmail},
		{name: "display name form is rejected", input: "Gayley <gayley@rub.edu.bt>", wantErr: ErrInvalidEmail},
		{name: "space inside", input: "gay ley@rub.edu.bt", wantErr: ErrInvalidEmail},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeEmail(tt.input)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  error
	}{
		{name: "7 characters is too short", password: "abc1234", wantErr: ErrPasswordTooShort},
		{name: "empty is too short", password: "", wantErr: ErrPasswordTooShort},
		{name: "8 characters is the minimum", password: "abcd1234"},
		{name: "72 bytes is the bcrypt maximum", password: strings.Repeat("a", 72)},
		{name: "73 bytes is too long", password: strings.Repeat("a", 73), wantErr: ErrPasswordTooLong},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestNewUser(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		fullName string
		password string
		wantErr  error
	}{
		{name: "valid user", email: "gayley@rub.edu.bt", fullName: "Gayley Choden", password: "secret123"},
		{name: "invalid email", email: "not-an-email", fullName: "Gayley Choden", password: "secret123", wantErr: ErrInvalidEmail},
		{name: "blank name", email: "gayley@rub.edu.bt", fullName: "   ", password: "secret123", wantErr: ErrInvalidName},
		{name: "short password", email: "gayley@rub.edu.bt", fullName: "Gayley Choden", password: "short", wantErr: ErrPasswordTooShort},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := NewUser("user-1", tt.email, tt.fullName, tt.password, fixedNow)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, "user-1", u.ID)
			assert.Equal(t, "Gayley Choden", u.FullName)
			assert.Equal(t, []Role{RoleStudent}, u.Roles, "new accounts get the STUDENT role")
			assert.True(t, u.Active)
			assert.Equal(t, fixedNow, u.CreatedAt)
			assert.NotEqual(t, tt.password, u.PasswordHash, "the password is never stored as plain text")
			assert.True(t, strings.HasPrefix(u.PasswordHash, "$2"), "the hash is a bcrypt hash")
		})
	}
}

func TestCheckPassword(t *testing.T) {
	u, err := NewUser("user-1", "gayley@rub.edu.bt", "Gayley Choden", "secret123", fixedNow)
	require.NoError(t, err)

	inactive := u
	inactive.Active = false

	tests := []struct {
		name     string
		user     User
		password string
		wantErr  error
	}{
		{name: "correct password", user: u, password: "secret123"},
		{name: "wrong password", user: u, password: "secret124", wantErr: ErrInvalidCredentials},
		{name: "empty password", user: u, password: "", wantErr: ErrInvalidCredentials},
		{name: "inactive user", user: inactive, password: "secret123", wantErr: ErrInvalidCredentials},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.CheckPassword(tt.password)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			assert.NoError(t, err)
		})
	}
}
