package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Short names for the five roles, used as columns in the table below.
const (
	S  = Student
	CH = CategoryHead
	SF = Staff
	AD = SystemAdmin
	PR = President
)

// expected is the role table from docs/design/api-routes.md, one row per
// action. Every role not listed for an action must be denied.
var expected = map[string][]Role{
	ActionMeRead:               {S, CH, SF, AD, PR},
	ActionCategoryList:         {S, CH, SF, AD, PR},
	ActionLocationList:         {S, CH, SF, AD, PR},
	ActionNotificationList:     {S, CH, SF, AD, PR},
	ActionPhotoUpload:          {S},
	ActionPhotoRead:            {S, CH, SF, AD},
	ActionReportSubmit:         {S},
	ActionReportRead:           {S, CH, SF, AD},
	ActionReportListOwn:        {S},
	ActionReportTimelineRead:   {S, CH, SF, AD},
	ActionReportStatusRead:     {S, CH, SF, AD},
	ActionReportStatusChange:   {CH, SF},
	ActionReportReassign:       {CH, SF},
	ActionReportAssignmentRead: {CH, SF, AD},
	ActionReportNoteAdd:        {CH, SF},
	ActionReportNoteRead:       {CH, SF, AD},
	ActionReportFeedbackSubmit: {S},
	ActionReportFeedbackRead:   {S, CH, AD},
	ActionReportAuditRead:      {AD},
	ActionDepartmentList:       {CH, AD},
	ActionRoutingRuleRead:      {AD},
	ActionRoutingRuleWrite:     {AD},
	ActionSLAPolicyRead:        {AD},
	ActionSLAPolicyWrite:       {AD},
	ActionDashboardRead:        {AD, PR},
}

var allRoles = []Role{S, CH, SF, AD, PR}

// TestEveryRoleAgainstEveryAction checks all 5 roles x 25 actions.
func TestEveryRoleAgainstEveryAction(t *testing.T) {
	p := DefaultPolicy()
	require.Len(t, p.Actions(), len(expected), "the policy has exactly the actions in the route table")

	for action, allowedRoles := range expected {
		for _, role := range allRoles {
			want := false
			for _, r := range allowedRoles {
				if r == role {
					want = true
				}
			}
			t.Run(role.String()+" "+action, func(t *testing.T) {
				d, err := p.Check(role, action, "report")
				require.NoError(t, err)
				assert.Equal(t, want, d.Allowed)
				if want {
					assert.Empty(t, d.Reason)
				} else {
					assert.NotEmpty(t, d.Reason, "a denial always says why")
				}
			})
		}
	}
}

// TestRoleRules spells out the important rules one by one, so a reader does
// not have to read the whole table.
func TestRoleRules(t *testing.T) {
	p := DefaultPolicy()
	tests := []struct {
		name   string
		role   Role
		action string
		want   bool
	}{
		{"student can submit a report", S, ActionReportSubmit, true},
		{"student can read a report", S, ActionReportRead, true},
		{"student cannot change status", S, ActionReportStatusChange, false},
		{"student cannot see the dashboard", S, ActionDashboardRead, false},
		{"staff can change status", SF, ActionReportStatusChange, true},
		{"staff can add an internal note", SF, ActionReportNoteAdd, true},
		{"staff cannot see the dashboard", SF, ActionDashboardRead, false},
		{"staff cannot manage routing rules", SF, ActionRoutingRuleWrite, false},
		{"category head can reassign", CH, ActionReportReassign, true},
		{"category head can list departments", CH, ActionDepartmentList, true},
		{"category head cannot submit a report", CH, ActionReportSubmit, false},
		{"admin can manage routing rules", AD, ActionRoutingRuleWrite, true},
		{"admin can see the dashboard", AD, ActionDashboardRead, true},
		{"admin can read the audit log", AD, ActionReportAuditRead, true},
		{"admin cannot change status", AD, ActionReportStatusChange, false},
		{"president can see the dashboard", PR, ActionDashboardRead, true},
		{"president cannot change status", PR, ActionReportStatusChange, false},
		{"president cannot reassign", PR, ActionReportReassign, false},
		{"president cannot manage routing rules", PR, ActionRoutingRuleWrite, false},
		{"president cannot submit a report", PR, ActionReportSubmit, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := p.Check(tt.role, tt.action, "report")
			require.NoError(t, err)
			assert.Equal(t, tt.want, d.Allowed)
		})
	}
}

func TestDefaultDeny(t *testing.T) {
	p := DefaultPolicy()
	tests := []struct {
		name   string
		role   Role
		action string
	}{
		{"unknown action", AD, "report.delete"},
		{"action in the wrong case", S, "REPORT.SUBMIT"},
		{"unknown role number", Role(9), ActionMeRead},
		{"negative role number", Role(-1), ActionMeRead},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := p.Check(tt.role, tt.action, "report")
			require.NoError(t, err)
			assert.False(t, d.Allowed)
			assert.NotEmpty(t, d.Reason)
		})
	}
}

func TestCheckRejectsMissingInput(t *testing.T) {
	p := DefaultPolicy()
	tests := []struct {
		name     string
		role     Role
		action   string
		resource string
		wantErr  error
	}{
		{"role unspecified", RoleUnspecified, ActionMeRead, "report", ErrRoleRequired},
		{"empty action", S, "", "report", ErrActionRequired},
		{"blank action", S, "   ", "report", ErrActionRequired},
		{"empty resource", S, ActionMeRead, "", ErrResourceRequired},
		{"blank resource", S, ActionMeRead, "  ", ErrResourceRequired},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := p.Check(tt.role, tt.action, tt.resource)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestResourceWithIDIsAccepted(t *testing.T) {
	d, err := DefaultPolicy().Check(S, ActionReportRead, "report:6f1c2b9e-0000-4000-8000-000000000001")
	require.NoError(t, err)
	assert.True(t, d.Allowed, "ownership is checked by the caller, not here")
}

func TestNewPolicyCopiesItsInput(t *testing.T) {
	rules := map[string][]Role{ActionMeRead: {S}}
	p := NewPolicy(rules)
	rules[ActionMeRead] = append(rules[ActionMeRead], PR)

	d, err := p.Check(PR, ActionMeRead, "me")
	require.NoError(t, err)
	assert.False(t, d.Allowed, "changing the input map later must not change the policy")
}

func TestRoleString(t *testing.T) {
	tests := []struct {
		role Role
		want string
	}{
		{RoleUnspecified, "ROLE_UNSPECIFIED"},
		{S, "STUDENT"},
		{CH, "CATEGORY_HEAD"},
		{SF, "STAFF"},
		{AD, "SYSTEM_ADMIN"},
		{PR, "PRESIDENT"},
		{Role(9), "ROLE_UNKNOWN"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, tt.role.String())
	}
}
