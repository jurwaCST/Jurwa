package domain

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMatchDefaultRules(t *testing.T) {
	tests := []struct {
		name    string
		query   Query
		want    string
		wantErr error
	}{
		{name: "hostel goes to DSA", query: Query{Category: Hostel, Building: "Hostel B"}, want: "DSA"},
		{name: "academic block goes to DAA", query: Query{Category: AcademicBlock, Building: "Block 2"}, want: "DAA"},
		{name: "others goes to SSO", query: Query{Category: Others}, want: "SSO"},
		{name: "unspecified category is invalid", query: Query{Category: ""}, wantErr: ErrInvalidCategory},
		{name: "unknown category is invalid", query: Query{Category: "LIBRARY"}, wantErr: ErrInvalidCategory},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Match(DefaultRules(), tt.query)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got.TargetPosition)
		})
	}
}

func TestMatchSelection(t *testing.T) {
	categoryRule := Rule{ID: "1", Category: Hostel, Building: "", TargetPosition: "DSA", Priority: 100, Active: true}
	buildingRule := Rule{ID: "2", Category: Hostel, Building: "Hostel B", TargetPosition: "SSO", Priority: 100, Active: true}
	inactiveRule := Rule{ID: "3", Category: Hostel, Building: "Hostel C", TargetPosition: "SSO", Priority: 100, Active: false}
	urgentRule := Rule{ID: "4", Category: Others, Building: "", TargetPosition: "ESTATE", Priority: 10, Active: true}
	normalRule := Rule{ID: "5", Category: Others, Building: "", TargetPosition: "SSO", Priority: 100, Active: true}

	tests := []struct {
		name    string
		rules   []Rule
		query   Query
		wantID  string
		wantErr error
	}{
		{name: "building rule beats category rule", rules: []Rule{categoryRule, buildingRule}, query: Query{Category: Hostel, Building: "Hostel B"}, wantID: "2"},
		{name: "other building falls back to category rule", rules: []Rule{categoryRule, buildingRule}, query: Query{Category: Hostel, Building: "Hostel A"}, wantID: "1"},
		{name: "inactive rule is ignored", rules: []Rule{categoryRule, inactiveRule}, query: Query{Category: Hostel, Building: "Hostel C"}, wantID: "1"},
		{name: "lower priority number wins", rules: []Rule{normalRule, urgentRule}, query: Query{Category: Others}, wantID: "4"},
		{name: "rules of other categories are ignored", rules: []Rule{urgentRule}, query: Query{Category: Hostel}, wantErr: ErrRuleNotFound},
		{name: "no rules at all", rules: nil, query: Query{Category: Hostel}, wantErr: ErrRuleNotFound},
		{name: "only inactive rules", rules: []Rule{inactiveRule}, query: Query{Category: Hostel, Building: "Hostel C"}, wantErr: ErrRuleNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Match(tt.rules, tt.query)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantID, got.ID)
		})
	}
}

// Routing uses only category and location (decision 1, FR-11). This test fails if
// anyone adds a description, or any other field, to the routing input.
func TestQueryHasNoDescription(t *testing.T) {
	typ := reflect.TypeOf(Query{})
	var fields []string
	for i := 0; i < typ.NumField(); i++ {
		fields = append(fields, typ.Field(i).Name)
	}
	assert.Equal(t, []string{"Category", "Building"}, fields)
}

func TestRuleValidate(t *testing.T) {
	tests := []struct {
		name    string
		rule    Rule
		wantErr error
	}{
		{name: "valid category rule", rule: Rule{Category: Hostel, TargetPosition: "DSA", Priority: 100}},
		{name: "valid building rule", rule: Rule{Category: AcademicBlock, Building: "Block 2", TargetPosition: "HOD_IT", Priority: 50}},
		{name: "missing category", rule: Rule{TargetPosition: "DSA", Priority: 100}, wantErr: ErrInvalidCategory},
		{name: "unknown category", rule: Rule{Category: "LIBRARY", TargetPosition: "DSA", Priority: 100}, wantErr: ErrInvalidCategory},
		{name: "missing target position", rule: Rule{Category: Hostel, Priority: 100}, wantErr: ErrInvalidRule},
		{name: "blank target position", rule: Rule{Category: Hostel, TargetPosition: "  ", Priority: 100}, wantErr: ErrInvalidRule},
		{name: "negative priority", rule: Rule{Category: Hostel, TargetPosition: "DSA", Priority: -1}, wantErr: ErrInvalidRule},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.rule.Validate()
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestDefaultRules(t *testing.T) {
	rules := DefaultRules()
	require.Len(t, rules, 3)
	for _, r := range rules {
		assert.NoError(t, r.Validate())
		assert.True(t, r.Active)
		assert.Empty(t, r.Building)
		assert.Equal(t, DefaultPriority, r.Priority)
	}
}
