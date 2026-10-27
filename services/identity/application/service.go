// Package application holds the identity use cases: Register, Login and
// ValidateToken. It does not know whether it is called over gRPC or a test.
package application

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/jurwaCST/Jurwa/services/identity/domain"
	"github.com/jurwaCST/Jurwa/services/identity/ports"
)

// dummyHash is compared against when the email is unknown, so a wrong email
// takes as long as a wrong password and cannot be detected by timing.
var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("dummy-password"), bcrypt.DefaultCost)

type Service struct {
	users  ports.UserRepository
	tokens *domain.TokenIssuer
	newID  func() string
	now    func() time.Time
}

func NewService(users ports.UserRepository, tokens *domain.TokenIssuer) *Service {
	return &Service{
		users:  users,
		tokens: tokens,
		newID:  func() string { return uuid.NewString() },
		now:    time.Now,
	}
}

// LoginResult is returned by a successful Login.
type LoginResult struct {
	AccessToken string
	ExpiresAt   time.Time
	User        domain.User
}

// Register creates a Student account.
func (s *Service) Register(ctx context.Context, email, fullName, password string) (domain.User, error) {
	u, err := domain.NewUser(s.newID(), email, fullName, password, s.now().UTC())
	if err != nil {
		return domain.User{}, err
	}
	if err := s.users.Create(ctx, u); err != nil {
		return domain.User{}, err
	}
	return u, nil
}

// Login checks the email and password and returns a signed token. A wrong
// email and a wrong password give the same error.
func (s *Service) Login(ctx context.Context, email, password string) (LoginResult, error) {
	e, err := domain.NormalizeEmail(email)
	if err != nil {
		return LoginResult{}, domain.ErrInvalidCredentials
	}
	u, err := s.users.FindByEmail(ctx, e)
	if errors.Is(err, domain.ErrUserNotFound) {
		_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(password))
		return LoginResult{}, domain.ErrInvalidCredentials
	}
	if err != nil {
		return LoginResult{}, err
	}
	if err := u.CheckPassword(password); err != nil {
		return LoginResult{}, err
	}
	token, expiresAt, err := s.tokens.Issue(u)
	if err != nil {
		return LoginResult{}, err
	}
	return LoginResult{AccessToken: token, ExpiresAt: expiresAt, User: u}, nil
}

// ValidateToken checks a token and returns the current user it belongs to.
func (s *Service) ValidateToken(ctx context.Context, token string) (domain.User, error) {
	claims, err := s.tokens.Parse(token)
	if err != nil {
		return domain.User{}, domain.ErrInvalidToken
	}
	u, err := s.users.FindByID(ctx, claims.UserID)
	if errors.Is(err, domain.ErrUserNotFound) || (err == nil && !u.Active) {
		return domain.User{}, domain.ErrInvalidToken
	}
	if err != nil {
		return domain.User{}, err
	}
	return u, nil
}
