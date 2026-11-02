package domain

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestValidate(t *testing.T) {
	tests := []struct {
		name    string
		req     Request
		wantErr error
	}{
		{name: "hostel", req: Request{Category: Hostel, Building: "Hostel B"}},
		{name: "academic block", req: Request{Category: AcademicBlock, Building: "Block 2"}},
		{name: "others", req: Request{Category: Others, Building: "Football Ground"}},
		{name: "with report id", req: Request{ReportID: "r1", Category: Others, Building: "Gate"}},
		{name: "unspecified category", req: Request{Building: "Hostel B"}, wantErr: ErrInvalidCategory},
		{name: "unknown category", req: Request{Category: "LIBRARY", Building: "Hostel B"}, wantErr: ErrInvalidCategory},
		{name: "empty location", req: Request{Category: Hostel}, wantErr: ErrInvalidLocation},
		{name: "blank location", req: Request{Category: Hostel, Building: "   "}, wantErr: ErrInvalidLocation},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestDecide(t *testing.T) {
	tests := []struct {
		name     string
		category Category
		rule     Rule
		holder   string
		wantErr  error
	}{
		{name: "hostel goes to DSA", category: Hostel, rule: Rule{ID: "rule-1", TargetPosition: "DSA"}, holder: "user-dsa"},
		{name: "academic block goes to DAA", category: AcademicBlock, rule: Rule{ID: "rule-2", TargetPosition: "DAA"}, holder: "user-daa"},
		{name: "others goes to SSO", category: Others, rule: Rule{ID: "rule-3", TargetPosition: "SSO"}, holder: "user-sso"},
		{name: "position without holder", category: Hostel, rule: Rule{ID: "rule-1", TargetPosition: "DSA"}, holder: "", wantErr: ErrNoHolder},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := Request{ReportID: "r1", Category: tt.category, Building: "B"}
			got, err := Decide(req, tt.rule, tt.holder)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, Decision{
				ReportID: "r1", Category: tt.category, CategoryHeadID: tt.holder,
				TargetPosition: tt.rule.TargetPosition, RuleID: tt.rule.ID,
			}, got)
		})
	}
}

// Routing uses only category and location (decision 1, FR-11). This test fails if
// anyone adds a description, or any other field, to the routing input.
func TestRequestHasNoDescription(t *testing.T) {
	typ := reflect.TypeOf(Request{})
	var fields []string
	for i := 0; i < typ.NumField(); i++ {
		fields = append(fields, typ.Field(i).Name)
	}
	assert.Equal(t, []string{"ReportID", "Category", "Building"}, fields)
}
