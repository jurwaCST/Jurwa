package application

import (
	"context"
	"fmt"

	"github.com/jurwaCST/Jurwa/services/api-gateway/domain"
	"github.com/jurwaCST/Jurwa/services/api-gateway/ports"
)

// Auth checks who the caller is (identity) and what they may do (authorization).
type Auth struct {
	identity ports.Identity
	authz    ports.Authorizer
}

func NewAuth(identity ports.Identity, authz ports.Authorizer) *Auth {
	return &Auth{identity: identity, authz: authz}
}

// Authenticate turns an Authorization header into a Principal. It returns
// domain.ErrMissingToken, domain.ErrMalformedToken or the identity error
// (Unauthenticated for a bad or expired token).
func (a *Auth) Authenticate(ctx context.Context, header string) (domain.Principal, error) {
	token, err := domain.BearerToken(header)
	if err != nil {
		return domain.Principal{}, err
	}
	u, err := a.identity.ValidateToken(ctx, token)
	if err != nil {
		return domain.Principal{}, err
	}
	p := domain.Principal{UserID: u.GetId(), Email: u.GetEmail()}
	for _, r := range u.GetRoles() {
		p.Roles = append(p.Roles, r.String())
	}
	return p, nil
}

// Authorize returns nil when any of p's roles may do action on resource,
// otherwise domain.ErrForbidden.
func (a *Auth) Authorize(ctx context.Context, p domain.Principal, action, resource string) error {
	for _, role := range p.Roles {
		ok, err := a.authz.CheckPermission(ctx, role, action, resource)
		if err != nil {
			return err
		}
		if ok {
			return nil
		}
	}
	return fmt.Errorf("%w: %s on %s", domain.ErrForbidden, action, resource)
}
