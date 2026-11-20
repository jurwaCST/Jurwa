package application

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jurwaCST/Jurwa/services/identity/domain"
)

const demoPassword = "demo-pass-123"

// The Category Head ids must match services/department-directory/domain
// (demoHolderIDs), so a routed report lands on a user who can log in.
const (
	dsaID = "00000000-0000-4000-8000-000000000001"
	daaID = "00000000-0000-4000-8000-000000000002"
	ssoID = "00000000-0000-4000-8000-000000000003"
)

func TestSeedDemoUsersCreatesOneUserPerDemoAccount(t *testing.T) {
	svc, _ := newTestService(t)
	ctx := context.Background()

	created, err := svc.SeedDemoUsers(ctx, demoPassword)
	require.NoError(t, err)
	assert.Equal(t, 24, created)

	want := []struct {
		id    string
		email string
		role  domain.Role
	}{
		{dsaID, "dsa.demo@jurwa.example", domain.RoleCategoryHead},
		{daaID, "daa.demo@jurwa.example", domain.RoleCategoryHead},
		{ssoID, "sso.demo@jurwa.example", domain.RoleCategoryHead},
		{"00000000-0000-4000-8000-000000000010", "staff.demo@jurwa.example", domain.RoleStaff},
		{"00000000-0000-4000-8000-000000000020", "admin.demo@jurwa.example", domain.RoleSystemAdmin},
		{"00000000-0000-4000-8000-000000000030", "president.demo@jurwa.example", domain.RolePresident},
		// SCRUM-158: the HoDs get Academic Block reports, the councillors get
		// Hostel and Others reports from the SSO. They fix them, so they are staff.
		{"00000000-0000-4000-8000-000000000041", "hod.arch@jurwa.example", domain.RoleStaff},
		{"00000000-0000-4000-8000-000000000042", "hod.civil@jurwa.example", domain.RoleStaff},
		{"00000000-0000-4000-8000-000000000043", "hod.ct@jurwa.example", domain.RoleStaff},
		{"00000000-0000-4000-8000-000000000044", "hod.eee@jurwa.example", domain.RoleStaff},
		{"00000000-0000-4000-8000-000000000045", "hod.mech@jurwa.example", domain.RoleStaff},
		{"00000000-0000-4000-8000-000000000046", "hod.sh@jurwa.example", domain.RoleStaff},
		{"00000000-0000-4000-8000-000000000051", "councillor.ha@jurwa.example", domain.RoleStaff},
		{"00000000-0000-4000-8000-000000000052", "councillor.hb@jurwa.example", domain.RoleStaff},
		{"00000000-0000-4000-8000-000000000053", "councillor.hc@jurwa.example", domain.RoleStaff},
		{"00000000-0000-4000-8000-000000000054", "councillor.hd@jurwa.example", domain.RoleStaff},
		{"00000000-0000-4000-8000-000000000055", "councillor.he@jurwa.example", domain.RoleStaff},
		{"00000000-0000-4000-8000-000000000056", "councillor.hf@jurwa.example", domain.RoleStaff},
		{"00000000-0000-4000-8000-000000000057", "councillor.nk@jurwa.example", domain.RoleStaff},
		{"00000000-0000-4000-8000-000000000058", "councillor.rka@jurwa.example", domain.RoleStaff},
		{"00000000-0000-4000-8000-000000000059", "councillor.rkb@jurwa.example", domain.RoleStaff},
		{"00000000-0000-4000-8000-000000000060", "councillor.games@jurwa.example", domain.RoleStaff},
		{"00000000-0000-4000-8000-000000000061", "councillor.mess@jurwa.example", domain.RoleStaff},
		{"00000000-0000-4000-8000-000000000062", "councillor.electrical@jurwa.example", domain.RoleStaff},
	}
	for _, w := range want {
		u, err := svc.users.FindByID(ctx, w.id)
		require.NoError(t, err, w.email)
		assert.Equal(t, w.email, u.Email)
		assert.Equal(t, []domain.Role{w.role}, u.Roles, w.email)
		assert.True(t, u.Active)
		assert.NotEmpty(t, u.FullName)
		// Only the bcrypt hash is stored.
		assert.NotContains(t, u.PasswordHash, demoPassword)
		assert.NoError(t, u.CheckPassword(demoPassword))
	}
	assert.Len(t, DemoUsers, len(want), "the list in the test and in the code must match")
}

// SCRUM-158: an HoD or councillor is shown by their position, never as "Staff".
func TestDemoStaffAreNamedByPosition(t *testing.T) {
	names := map[string]string{}
	for _, d := range DemoUsers {
		names[d.Email] = d.FullName
	}
	assert.Equal(t, "Block Councillor (HA)", names["councillor.ha@jurwa.example"])
	assert.Equal(t, "Games and Health Councillor", names["councillor.games@jurwa.example"])
	assert.Equal(t, "Deputy Chief Councillor (Mess)", names["councillor.mess@jurwa.example"])
	assert.Equal(t, "Electrical Maintenance Councillor", names["councillor.electrical@jurwa.example"])
	assert.Equal(t, "HoD, Computing Technologies", names["hod.ct@jurwa.example"])
	assert.Equal(t, "HoD, Mechanical Engineering (acting)", names["hod.mech@jurwa.example"])
}

func TestSeedDemoUsersTwiceCreatesNoDuplicates(t *testing.T) {
	svc, _ := newTestService(t)
	ctx := context.Background()

	_, err := svc.SeedDemoUsers(ctx, demoPassword)
	require.NoError(t, err)
	created, err := svc.SeedDemoUsers(ctx, demoPassword)
	require.NoError(t, err)
	assert.Equal(t, 0, created, "second run creates nothing")

	u, err := svc.users.FindByEmail(ctx, "dsa.demo@jurwa.example")
	require.NoError(t, err)
	assert.Equal(t, dsaID, u.ID)
}

func TestSeedDemoUsersNeedsAPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  error
	}{
		{name: "missing password", password: "", wantErr: ErrDemoPasswordMissing},
		{name: "blank password", password: "   ", wantErr: ErrDemoPasswordMissing},
		{name: "too short", password: "1234567", wantErr: domain.ErrPasswordTooShort},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, _ := newTestService(t)
			ctx := context.Background()
			created, err := svc.SeedDemoUsers(ctx, tt.password)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, 0, created)
			_, err = svc.users.FindByID(ctx, dsaID)
			assert.ErrorIs(t, err, domain.ErrUserNotFound, "nothing is created")
		})
	}
}

func TestSeedDemoUsersRefusesAnEmailHeldByAnotherAccount(t *testing.T) {
	svc, _ := newTestService(t)
	ctx := context.Background()
	// Someone registered the demo address first, as a Student.
	_, err := svc.Register(ctx, "dsa.demo@jurwa.example", "Not The DSA", "secret123")
	require.NoError(t, err)

	_, err = svc.SeedDemoUsers(ctx, demoPassword)
	assert.ErrorIs(t, err, ErrDemoEmailTaken)
}

func TestSeededUserCanLogInWithTheRightRoleAndID(t *testing.T) {
	svc, _ := newTestService(t)
	ctx := context.Background()
	_, err := svc.SeedDemoUsers(ctx, demoPassword)
	require.NoError(t, err)

	res, err := svc.Login(ctx, "dsa.demo@jurwa.example", demoPassword)
	require.NoError(t, err)
	assert.Equal(t, dsaID, res.User.ID)
	assert.Equal(t, []domain.Role{domain.RoleCategoryHead}, res.User.Roles)

	claims, err := svc.tokens.Parse(res.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, dsaID, claims.UserID)
	assert.Equal(t, []domain.Role{domain.RoleCategoryHead}, claims.Roles)

	_, err = svc.Login(ctx, "dsa.demo@jurwa.example", "wrong-password")
	assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
}

func TestWithoutTheSeedDemoUsersCannotLogIn(t *testing.T) {
	svc, _ := newTestService(t)
	_, err := svc.Login(context.Background(), "dsa.demo@jurwa.example", demoPassword)
	assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
}

func TestRegisterStillGivesOnlyStudentAfterTheSeed(t *testing.T) {
	svc, _ := newTestService(t)
	ctx := context.Background()
	_, err := svc.SeedDemoUsers(ctx, demoPassword)
	require.NoError(t, err)

	// Register has no role input at all (RegisterRequest has no role field),
	// so a new account is always a Student, whatever else exists.
	u, err := svc.Register(ctx, "pema@rub.edu.bt", "Pema", "secret123")
	require.NoError(t, err)
	assert.Equal(t, []domain.Role{domain.RoleStudent}, u.Roles)
}
