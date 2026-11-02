package domain

import (
	"errors"
	"slices"
	"strings"
)

// Errors for bad input. The gRPC adapter maps them to InvalidArgument.
var (
	ErrRoleRequired     = errors.New("role is required")
	ErrActionRequired   = errors.New("action is required")
	ErrResourceRequired = errors.New("resource is required")
)

// Role of a user. The numbers match jurwa.common.v1.Role, but the domain
// does not import the generated proto code.
type Role int32

const (
	RoleUnspecified Role = 0
	Student         Role = 1
	CategoryHead    Role = 2
	Staff           Role = 3
	SystemAdmin     Role = 4
	President       Role = 5 // read-only dashboard
)

var roleNames = map[Role]string{
	RoleUnspecified: "ROLE_UNSPECIFIED",
	Student:         "STUDENT",
	CategoryHead:    "CATEGORY_HEAD",
	Staff:           "STAFF",
	SystemAdmin:     "SYSTEM_ADMIN",
	President:       "PRESIDENT",
}

// String returns the role name used in the proto enum.
func (r Role) String() string {
	if name, ok := roleNames[r]; ok {
		return name
	}
	return "ROLE_UNKNOWN"
}

// Actions, one per route in docs/design/api-routes.md. The api-gateway sends
// these strings. The first five match services/api-gateway/domain/auth.go.
const (
	ActionMeRead               = "me.read"
	ActionCategoryList         = "category.list"
	ActionLocationList         = "location.list"
	ActionNotificationList     = "notification.list"
	ActionPhotoUpload          = "photo.upload"
	ActionPhotoRead            = "photo.read"
	ActionReportSubmit         = "report.submit"
	ActionReportRead           = "report.read"
	ActionReportListOwn        = "report.list_own"
	ActionReportTimelineRead   = "report.timeline.read"
	ActionReportStatusRead     = "report.status.read"
	ActionReportStatusChange   = "report.status.change"
	ActionReportReassign       = "report.reassign"
	ActionReportAssignmentRead = "report.assignments.read"
	ActionReportNoteAdd        = "report.notes.add"
	ActionReportNoteRead       = "report.notes.read"
	ActionReportFeedbackSubmit = "report.feedback.submit"
	ActionReportFeedbackRead   = "report.feedback.read"
	ActionReportAuditRead      = "report.audit.read"
	ActionDepartmentList       = "department.list"
	ActionRoutingRuleRead      = "routing_rule.read"
	ActionRoutingRuleWrite     = "routing_rule.write"
	ActionSLAPolicyRead        = "sla_policy.read"
	ActionSLAPolicyWrite       = "sla_policy.write"
	ActionDashboardRead        = "dashboard.read"
)

// Decision is the answer to a permission check.
type Decision struct {
	Allowed bool
	Reason  string // why it was denied, "" when allowed
}

// Policy maps each action to the roles allowed to do it. Anything not in the
// map is denied (default deny).
type Policy struct {
	rules map[string][]Role
}

// NewPolicy builds a policy from a copy of rules, so later changes to the
// map do not change the policy.
func NewPolicy(rules map[string][]Role) Policy {
	cp := make(map[string][]Role, len(rules))
	for action, roles := range rules {
		cp[action] = slices.Clone(roles)
	}
	return Policy{rules: cp}
}

// DefaultPolicy is the role table from docs/design/api-routes.md.
func DefaultPolicy() Policy {
	anyone := []Role{Student, CategoryHead, Staff, SystemAdmin, President}
	reportReaders := []Role{Student, CategoryHead, Staff, SystemAdmin}
	return NewPolicy(map[string][]Role{
		ActionMeRead:               anyone,
		ActionCategoryList:         anyone,
		ActionLocationList:         anyone,
		ActionNotificationList:     anyone,
		ActionPhotoUpload:          {Student},
		ActionPhotoRead:            reportReaders,
		ActionReportSubmit:         {Student},
		ActionReportRead:           reportReaders,
		ActionReportListOwn:        {Student},
		ActionReportTimelineRead:   reportReaders,
		ActionReportStatusRead:     reportReaders,
		ActionReportStatusChange:   {CategoryHead, Staff},
		ActionReportReassign:       {CategoryHead, Staff},
		ActionReportAssignmentRead: {CategoryHead, Staff, SystemAdmin},
		ActionReportNoteAdd:        {CategoryHead, Staff},
		ActionReportNoteRead:       {CategoryHead, Staff, SystemAdmin},
		ActionReportFeedbackSubmit: {Student},
		ActionReportFeedbackRead:   {Student, CategoryHead, SystemAdmin},
		ActionReportAuditRead:      {SystemAdmin},
		ActionDepartmentList:       {CategoryHead, SystemAdmin},
		ActionRoutingRuleRead:      {SystemAdmin},
		ActionRoutingRuleWrite:     {SystemAdmin},
		ActionSLAPolicyRead:        {SystemAdmin},
		ActionSLAPolicyWrite:       {SystemAdmin},
		ActionDashboardRead:        {SystemAdmin, President},
	})
}

// Actions returns every action the policy knows, sorted.
func (p Policy) Actions() []string {
	out := make([]string, 0, len(p.rules))
	for action := range p.rules {
		out = append(out, action)
	}
	slices.Sort(out)
	return out
}

// Check says whether role may do action on resource. Missing input is an
// error. An unknown role or action is denied, not an error.
//
// resource is required but not used for the decision: whether a Student owns
// a report is checked by the caller, which has the report data.
func (p Policy) Check(role Role, action, resource string) (Decision, error) {
	if role == RoleUnspecified {
		return Decision{}, ErrRoleRequired
	}
	if strings.TrimSpace(action) == "" {
		return Decision{}, ErrActionRequired
	}
	if strings.TrimSpace(resource) == "" {
		return Decision{}, ErrResourceRequired
	}
	if _, known := roleNames[role]; !known {
		return Decision{Reason: "unknown role"}, nil
	}
	roles, known := p.rules[action]
	if !known {
		return Decision{Reason: "unknown action " + action}, nil
	}
	if !slices.Contains(roles, role) {
		return Decision{Reason: role.String() + " may not " + action}, nil
	}
	return Decision{Allowed: true}, nil
}
