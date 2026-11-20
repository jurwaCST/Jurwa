package spec

// SCRUM-158: routing by position. Academic Block reports go to the HoD of the
// department the student picks; Hostel and Others reports go to the SSO, who
// hands them to a councillor. A late report goes up to the DAA or the DSA.
// These tests join the data of five services, so a name changed in one of them
// (a department, a position code, a demo account) fails here.

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	directory "github.com/jurwaCST/Jurwa/services/department-directory/domain"
	identity "github.com/jurwaCST/Jurwa/services/identity/application"
	identitydomain "github.com/jurwaCST/Jurwa/services/identity/domain"
	locations "github.com/jurwaCST/Jurwa/services/location-taxonomy/adapters/out/memory"
	location "github.com/jurwaCST/Jurwa/services/location-taxonomy/domain"
	reassignment "github.com/jurwaCST/Jurwa/services/reassignment/domain"
	rules "github.com/jurwaCST/Jurwa/services/routing-rules/domain"
)

// demoUsers maps a demo account id to the account.
func demoUsers() map[string]identity.DemoUser {
	out := map[string]identity.DemoUser{}
	for _, u := range identity.DemoUsers {
		out[u.ID] = u
	}
	return out
}

func heldPosition(t *testing.T, code string) directory.Position {
	t.Helper()
	p, err := directory.FindPosition(directory.WithDemoHolders(directory.DefaultPositions()), code)
	require.NoError(t, err, "position %s exists", code)
	require.NotEmpty(t, p.HolderUserID, "position %s has a holder", code)
	return p
}

// Every active place a student can pick routes to a position someone holds, and
// that someone is a demo account who can log in.
func TestEveryPlaceReachesSomeoneWhoCanLogIn(t *testing.T) {
	users := demoUsers()
	departments := map[string]string{} // department name to code
	for _, d := range directory.DefaultDepartments() {
		departments[d.Name] = d.Code
	}

	for _, e := range location.ActiveOnly(locations.SampleEntries()) {
		name := fmt.Sprintf("%s %s %s %s", e.Category, e.Location.Building, e.Location.Wing, e.Location.Area)
		rule, err := rules.Match(rules.DefaultRules(), rules.Query{Category: rules.Category(e.Category), Building: e.Location.Building})
		require.NoError(t, err, name)
		p := heldPosition(t, rule.TargetPosition)
		user, ok := users[p.HolderUserID]
		require.True(t, ok, "%s: holder of %s is a demo account", name, p.Code)

		switch e.Category {
		case location.CategoryAcademicBlock:
			assert.Equal(t, directory.KindHoD, p.Kind, name)
			assert.Equal(t, departments[e.Location.Building], p.DepartmentCode, "%s goes to its own department's HoD", name)
			assert.Equal(t, identitydomain.RoleStaff, user.Role, "an HoD fixes the report and cannot reassign it")
		default:
			assert.Equal(t, "SSO", p.Code, name)
			assert.Equal(t, identitydomain.RoleCategoryHead, user.Role, "the SSO hands reports out")
		}
	}
}

// A late report goes up to the DSA (from the SSO) or the DAA (from an HoD),
// both held by a Category Head demo account.
func TestEscalationGoesUpToTheDSAOrTheDAA(t *testing.T) {
	users := demoUsers()
	want := map[string]string{"SSO": "DSA"}
	for _, d := range directory.DefaultDepartments() {
		want["HOD_"+d.Code] = "DAA"
	}
	for routedTo, up := range want {
		assert.Equal(t, up, reassignment.EscalationPosition(routedTo), "routed to %s", routedTo)
		p := heldPosition(t, up)
		assert.Equal(t, identitydomain.RoleCategoryHead, users[p.HolderUserID].Role, "%s holder", up)
	}
}

// The SSO can hand a report to every councillor, and each councillor's account
// is named by their position, never "Staff".
func TestCouncillorsCanLogInUnderTheirPositionName(t *testing.T) {
	users := demoUsers()
	n := 0
	for _, p := range directory.WithDemoHolders(directory.DefaultPositions()) {
		if p.Kind != directory.KindCouncillor {
			continue
		}
		n++
		require.NotEmpty(t, p.HolderUserID, p.Code)
		user, ok := users[p.HolderUserID]
		require.True(t, ok, "holder of %s is a demo account", p.Code)
		assert.Equal(t, p.Title, user.FullName, p.Code)
		assert.Equal(t, identitydomain.RoleStaff, user.Role, p.Code)
	}
	assert.Equal(t, 12, n, "9 Block Councillors, Games and Health, Mess, Electrical Maintenance")
}
