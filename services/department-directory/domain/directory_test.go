package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The 6 academic departments of CST, as on https://cst.edu.bt/department/.
func TestDefaultDepartmentsAreCSTDepartments(t *testing.T) {
	assert.Equal(t, []Department{
		{Code: "ARCH", Name: "Architecture"},
		{Code: "CIVIL", Name: "Civil & Environmental Engineering"},
		{Code: "CT", Name: "Computing Technologies"},
		{Code: "EEE", Name: "Electrical & Electronics Engineering"},
		{Code: "MECH", Name: "Mechanical Engineering"},
		{Code: "SH", Name: "Science & Humanities"},
	}, DefaultDepartments())
}

func TestDefaultDepartments(t *testing.T) {
	depts := DefaultDepartments()
	require.Len(t, depts, 6)
	seen := map[string]bool{}
	for _, d := range depts {
		assert.NotEmpty(t, d.Code)
		assert.NotEmpty(t, d.Name)
		assert.False(t, seen[d.Code], "codes are unique")
		seen[d.Code] = true
	}
}

func TestFindHoD(t *testing.T) {
	depts, positions := DefaultDepartments(), DefaultPositions()

	for _, d := range depts {
		t.Run(d.Code, func(t *testing.T) {
			got, err := FindHoD(depts, positions, d.Code)
			require.NoError(t, err)
			assert.Equal(t, KindHoD, got.Kind)
			assert.Equal(t, d.Code, got.DepartmentCode)
			assert.Equal(t, "HOD_"+d.Code, got.Code)
		})
	}

	tests := []struct {
		name    string
		code    string
		wantErr error
	}{
		{name: "empty id", code: "", wantErr: ErrInvalidCode},
		{name: "blank id", code: "   ", wantErr: ErrInvalidCode},
		{name: "unknown department", code: "PHYSICS", wantErr: ErrNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := FindHoD(depts, positions, tt.code)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}

	// A department whose HoD position is missing is also not found.
	_, err := FindHoD(depts, nil, "CT")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestFindPosition(t *testing.T) {
	positions := DefaultPositions()

	tests := []struct {
		name     string
		code     string
		wantKind Kind
		wantErr  error
	}{
		// The three Category Heads (decision 2).
		{name: "DSA", code: "DSA", wantKind: KindCategoryHead},
		{name: "DAA", code: "DAA", wantKind: KindCategoryHead},
		{name: "SSO", code: "SSO", wantKind: KindCategoryHead},
		// Reassignment targets named in the proposal.
		{name: "SSO male", code: "SSO_MALE", wantKind: KindStaff},
		{name: "SSO female", code: "SSO_FEMALE", wantKind: KindStaff},
		{name: "block councillor", code: "BLOCK_COUNCILLOR", wantKind: KindCouncillor},
		{name: "maintenance councillor", code: "MAINTENANCE_COUNCILLOR", wantKind: KindCouncillor},
		{name: "deputy chief councillor mess", code: "DEPUTY_CHIEF_COUNCILLOR_MESS", wantKind: KindCouncillor},
		{name: "chief councillor", code: "CHIEF_COUNCILLOR", wantKind: KindCouncillor},
		{name: "code is case-insensitive", code: " dsa ", wantKind: KindCategoryHead},
		{name: "empty code", code: "", wantErr: ErrInvalidCode},
		{name: "unknown code", code: "PRINCIPAL", wantErr: ErrNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FindPosition(positions, tt.code)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantKind, got.Kind)
			assert.NotEmpty(t, got.Title)
			assert.True(t, got.Active)
		})
	}
}

func TestDefaultPositionsAreValid(t *testing.T) {
	seen := map[string]bool{}
	for _, p := range DefaultPositions() {
		assert.NoError(t, p.Validate(), p.Code)
		assert.False(t, seen[p.Code], "codes are unique: %s", p.Code)
		seen[p.Code] = true
	}
}

func TestPositionValidate(t *testing.T) {
	tests := []struct {
		name    string
		pos     Position
		wantErr error
	}{
		{name: "valid", pos: Position{Code: "DSA", Title: "Dean of Student Affairs", Kind: KindCategoryHead}},
		{name: "missing code", pos: Position{Title: "x", Kind: KindStaff}, wantErr: ErrInvalidCode},
		{name: "missing title", pos: Position{Code: "X", Kind: KindStaff}, wantErr: ErrInvalidPosition},
		{name: "unknown kind", pos: Position{Code: "X", Title: "x", Kind: "BOSS"}, wantErr: ErrInvalidPosition},
		{name: "HoD without department", pos: Position{Code: "HOD_X", Title: "x", Kind: KindHoD}, wantErr: ErrInvalidPosition},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.pos.Validate()
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestWithDemoHolders(t *testing.T) {
	positions := WithDemoHolders(DefaultPositions())
	for _, code := range []string{"DSA", "DAA", "SSO"} {
		p, err := FindPosition(positions, code)
		require.NoError(t, err)
		assert.NotEmpty(t, p.HolderUserID, code)
	}
	plain, err := FindPosition(DefaultPositions(), "DSA")
	require.NoError(t, err)
	assert.Empty(t, plain.HolderUserID, "the default list is not changed")
}

// Mechanical has no HoD on the staff directory, so its Programme Leader acts as
// HoD. The code stays HOD_MECH, so reassignment does not change.
func TestMechanicalProgrammeLeaderActsAsHoD(t *testing.T) {
	p, err := FindHoD(DefaultDepartments(), DefaultPositions(), "MECH")
	require.NoError(t, err)
	assert.Equal(t, "HOD_MECH", p.Code)
	assert.Equal(t, "Head of Department, Mechanical Engineering (Programme Leader, acting)", p.Title)

	ct, err := FindHoD(DefaultDepartments(), DefaultPositions(), "CT")
	require.NoError(t, err)
	assert.Equal(t, "Head of Department, Computing Technologies", ct.Title, "other departments keep the normal title")
}
