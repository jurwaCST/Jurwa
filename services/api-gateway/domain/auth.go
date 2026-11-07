package domain

import (
	"errors"
	"slices"
	"strings"
)

var (
	ErrMissingToken   = errors.New("missing bearer token")
	ErrMalformedToken = errors.New("authorization header must be \"Bearer <token>\"")
	ErrForbidden      = errors.New("not allowed")
)

// Role codes, the same strings as the proto enum names.
const (
	RoleStudent      = "STUDENT"
	RoleCategoryHead = "CATEGORY_HEAD"
	RoleStaff        = "STAFF"
	RoleSystemAdmin  = "SYSTEM_ADMIN"
)

// Actions checked against the authorization service, one per route.
const (
	ActionMeRead       = "me.read"
	ActionCategoryList = "category.list"
	ActionLocationList = "location.list"
	ActionReportSubmit = "report.submit"
	ActionReportRead   = "report.read"

	ActionReportNoteAdd  = "report.notes.add"
	ActionReportNoteRead = "report.notes.read"

	ActionReportListOwn      = "report.list_own"
	ActionReportTimelineRead = "report.timeline.read"

	ActionPhotoUpload = "photo.upload"
	ActionPhotoRead   = "photo.read"

	ActionReportStatusRead   = "report.status.read"
	ActionReportStatusChange = "report.status.change"

	ActionDashboardRead = "dashboard.read"
)

// Principal is the signed-in user, taken from a validated JWT.
type Principal struct {
	UserID string
	Email  string
	Roles  []string
}

// HasRole reports whether p has role.
func (p Principal) HasRole(role string) bool {
	return slices.Contains(p.Roles, role)
}

// BearerToken returns the token from an Authorization header value.
func BearerToken(header string) (string, error) {
	header = strings.TrimSpace(header)
	if header == "" {
		return "", ErrMissingToken
	}
	scheme, token, ok := strings.Cut(header, " ")
	if !ok || !strings.EqualFold(scheme, "Bearer") {
		return "", ErrMalformedToken
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return "", ErrMalformedToken
	}
	return token, nil
}

// CanReadReport applies "Student (own), Category Head, Staff, System
// Administrator" from docs/design/api-routes.md to one report.
func CanReadReport(p Principal, reporterID string) bool {
	if p.HasRole(RoleCategoryHead) || p.HasRole(RoleStaff) || p.HasRole(RoleSystemAdmin) {
		return true
	}
	return p.HasRole(RoleStudent) && p.UserID == reporterID
}
