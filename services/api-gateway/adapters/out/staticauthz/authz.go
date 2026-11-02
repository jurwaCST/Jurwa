// Package staticauthz is a stand-in Authorizer with the role table from
// docs/design/api-routes.md.
//
// TODO(SCRUM-38): replace it with a gRPC client for authorization
// CheckPermission (AUTHORIZATION_ADDR, default localhost:9017) once Gayley's
// authorization service is merged. Only cmd/server/main.go has to change.
package staticauthz

import (
	"context"
	"slices"

	"github.com/jurwaCST/Jurwa/services/api-gateway/domain"
)

var anyRole = []string{domain.RoleStudent, domain.RoleCategoryHead, domain.RoleStaff, domain.RoleSystemAdmin, "PRESIDENT"}

// table maps each action to the roles allowed to do it.
var table = map[string][]string{
	domain.ActionMeRead:       anyRole,
	domain.ActionCategoryList: anyRole,
	domain.ActionLocationList: anyRole,
	domain.ActionReportSubmit: {domain.RoleStudent},
	domain.ActionReportRead:   {domain.RoleStudent, domain.RoleCategoryHead, domain.RoleStaff, domain.RoleSystemAdmin},
}

// Authorizer implements ports.Authorizer from a fixed table.
type Authorizer struct{}

func New() Authorizer { return Authorizer{} }

// CheckPermission ignores resource: ownership ("Student (own)") is checked by
// the gateway with domain.CanReadReport.
func (Authorizer) CheckPermission(_ context.Context, role, action, _ string) (bool, error) {
	return slices.Contains(table[action], role), nil
}
