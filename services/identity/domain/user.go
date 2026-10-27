// Package domain holds the identity business rules: users, password rules and
// tokens. No gRPC or database code here.
package domain

import (
	"errors"
	"net/mail"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Errors returned by the domain. The gRPC adapter maps them to status codes.
var (
	ErrInvalidEmail       = errors.New("invalid email address")
	ErrInvalidName        = errors.New("full name is required")
	ErrPasswordTooShort   = errors.New("password must be at least 8 characters")
	ErrPasswordTooLong    = errors.New("password must be at most 72 bytes")
	ErrEmailTaken         = errors.New("email is already registered")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("wrong email or password")
	ErrInvalidToken       = errors.New("invalid or expired token")
)

const (
	MinPasswordLength = 8
	// bcrypt only uses the first 72 bytes, so longer passwords are rejected
	// instead of being silently cut.
	MaxPasswordBytes = 72
)

// Role matches jurwa.common.v1.Role. People reached by reassignment are Staff.
type Role string

const (
	RoleStudent      Role = "STUDENT"
	RoleCategoryHead Role = "CATEGORY_HEAD"
	RoleStaff        Role = "STAFF"
	RoleSystemAdmin  Role = "SYSTEM_ADMIN"
	RolePresident    Role = "PRESIDENT"
)

// User is an account. The password is only ever held as a bcrypt hash.
type User struct {
	ID           string
	Email        string
	FullName     string
	PasswordHash string
	Roles        []Role
	Active       bool
	CreatedAt    time.Time
}

// NormalizeEmail trims and lowercases an email address and checks that it is
// a plain address such as name@rub.edu.bt.
func NormalizeEmail(email string) (string, error) {
	e := strings.ToLower(strings.TrimSpace(email))
	addr, err := mail.ParseAddress(e)
	if err != nil || addr.Address != e || !strings.Contains(e[strings.LastIndex(e, "@")+1:], ".") {
		return "", ErrInvalidEmail
	}
	return e, nil
}

// ValidatePassword checks the password length rules.
func ValidatePassword(password string) error {
	if len([]rune(password)) < MinPasswordLength {
		return ErrPasswordTooShort
	}
	if len(password) > MaxPasswordBytes {
		return ErrPasswordTooLong
	}
	return nil
}

// NewUser validates the input and creates an active Student account with a
// bcrypt hash of the password.
func NewUser(id, email, fullName, password string, now time.Time) (User, error) {
	e, err := NormalizeEmail(email)
	if err != nil {
		return User{}, err
	}
	name := strings.TrimSpace(fullName)
	if name == "" {
		return User{}, ErrInvalidName
	}
	if err := ValidatePassword(password); err != nil {
		return User{}, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, err
	}
	return User{
		ID:           id,
		Email:        e,
		FullName:     name,
		PasswordHash: string(hash),
		Roles:        []Role{RoleStudent},
		Active:       true,
		CreatedAt:    now,
	}, nil
}

// CheckPassword returns ErrInvalidCredentials when the password is wrong or
// the account is inactive. It never says which of the two it was.
func (u User) CheckPassword(password string) error {
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		return ErrInvalidCredentials
	}
	if !u.Active {
		return ErrInvalidCredentials
	}
	return nil
}
