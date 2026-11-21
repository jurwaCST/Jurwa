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
// SCRUM-159: the accounts use the CST email format and new ids (...01NN), so a
// database that still holds the old @jurwa.example accounts is not touched.
const (
	dsaID = "00000000-0000-4000-8000-000000000101"
	daaID = "00000000-0000-4000-8000-000000000102"
	ssoID = "00000000-0000-4000-8000-000000000103"
)

func TestSeedDemoUsersCreatesOneUserPerDemoAccount(t *testing.T) {
	svc, _ := newTestService(t)
	ctx := context.Background()

	created, err := svc.SeedDemoUsers(ctx, demoPassword)
	require.NoError(t, err)
	assert.Equal(t, 25, created)

	want := []struct {
		id    string
		email string
		role  domain.Role
	}{
		{dsaID, "dsa.cst@rub.edu.bt", domain.RoleCategoryHead},
		{daaID, "daa.cst@rub.edu.bt", domain.RoleCategoryHead},
		{ssoID, "sso.cst@rub.edu.bt", domain.RoleCategoryHead},
		{"00000000-0000-4000-8000-000000000110", "staff.cst@rub.edu.bt", domain.RoleStaff},
		{"00000000-0000-4000-8000-000000000120", "admin.cst@rub.edu.bt", domain.RoleSystemAdmin},
		{"00000000-0000-4000-8000-000000000130", "president.cst@rub.edu.bt", domain.RolePresident},
		// SCRUM-158: the HoDs get Academic Block reports, the councillors get
		// Hostel and Others reports from the SSO. They fix them, so they are staff.
		{"00000000-0000-4000-8000-000000000141", "hod.arch.cst@rub.edu.bt", domain.RoleStaff},
		{"00000000-0000-4000-8000-000000000142", "hod.civil.cst@rub.edu.bt", domain.RoleStaff},
		{"00000000-0000-4000-8000-000000000143", "hod.ct.cst@rub.edu.bt", domain.RoleStaff},
		{"00000000-0000-4000-8000-000000000144", "hod.eee.cst@rub.edu.bt", domain.RoleStaff},
		{"00000000-0000-4000-8000-000000000145", "hod.mech.cst@rub.edu.bt", domain.RoleStaff},
		{"00000000-0000-4000-8000-000000000146", "hod.sh.cst@rub.edu.bt", domain.RoleStaff},
		{"00000000-0000-4000-8000-000000000151", "ha.cst@rub.edu.bt", domain.RoleStaff},
		{"00000000-0000-4000-8000-000000000152", "hb.cst@rub.edu.bt", domain.RoleStaff},
		{"00000000-0000-4000-8000-000000000153", "hc.cst@rub.edu.bt", domain.RoleStaff},
		{"00000000-0000-4000-8000-000000000154", "hd.cst@rub.edu.bt", domain.RoleStaff},
		{"00000000-0000-4000-8000-000000000155", "he.cst@rub.edu.bt", domain.RoleStaff},
		{"00000000-0000-4000-8000-000000000156", "hf.cst@rub.edu.bt", domain.RoleStaff},
		{"00000000-0000-4000-8000-000000000157", "nk.cst@rub.edu.bt", domain.RoleStaff},
		{"00000000-0000-4000-8000-000000000158", "rka.cst@rub.edu.bt", domain.RoleStaff},
		{"00000000-0000-4000-8000-000000000159", "rkb.cst@rub.edu.bt", domain.RoleStaff},
		{"00000000-0000-4000-8000-000000000160", "games.cst@rub.edu.bt", domain.RoleStaff},
		{"00000000-0000-4000-8000-000000000161", "mess.cst@rub.edu.bt", domain.RoleStaff},
		{"00000000-0000-4000-8000-000000000162", "electrical.cst@rub.edu.bt", domain.RoleStaff},
		// SCRUM-159: a demo student, in the student number format. Not a real student.
		{"00000000-0000-4000-8000-000000000199", "02240300.cst@rub.edu.bt", domain.RoleStudent},
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
	assert.Equal(t, "Block Councillor (HA)", names["ha.cst@rub.edu.bt"])
	assert.Equal(t, "Games and Health Councillor", names["games.cst@rub.edu.bt"])
	assert.Equal(t, "Deputy Chief Councillor (Mess)", names["mess.cst@rub.edu.bt"])
	assert.Equal(t, "Electrical Maintenance Councillor", names["electrical.cst@rub.edu.bt"])
	assert.Equal(t, "HoD, Computing Technologies", names["hod.ct.cst@rub.edu.bt"])
	assert.Equal(t, "HoD, Mechanical Engineering (acting)", names["hod.mech.cst@rub.edu.bt"])
	// SCRUM-159: the heads are named by post too, never "DSA (demo)".
	assert.Equal(t, "Dean of Student Affairs", names["dsa.cst@rub.edu.bt"])
	assert.Equal(t, "Dean of Academic Affairs", names["daa.cst@rub.edu.bt"])
	assert.Equal(t, "Student Service Officer", names["sso.cst@rub.edu.bt"])
	assert.Equal(t, "Student Service Officer (male)", names["staff.cst@rub.edu.bt"])
}

// SCRUM-159: every demo account is in the CST format, <name>.cst@rub.edu.bt.
func TestDemoEmailsUseTheCSTFormat(t *testing.T) {
	for _, d := range DemoUsers {
		assert.Regexp(t, `^[a-z0-9.]+\.cst@rub\.edu\.bt$`, d.Email)
	}
}

func TestSeedDemoUsersTwiceCreatesNoDuplicates(t *testing.T) {
	svc, _ := newTestService(t)
	ctx := context.Background()

	_, err := svc.SeedDemoUsers(ctx, demoPassword)
	require.NoError(t, err)
	created, err := svc.SeedDemoUsers(ctx, demoPassword)
	require.NoError(t, err)
	assert.Equal(t, 0, created, "second run creates nothing")

	u, err := svc.users.FindByEmail(ctx, "dsa.cst@rub.edu.bt")
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
	_, err := svc.Register(ctx, "dsa.cst@rub.edu.bt", "Not The DSA", "secret123")
	require.NoError(t, err)

	_, err = svc.SeedDemoUsers(ctx, demoPassword)
	assert.ErrorIs(t, err, ErrDemoEmailTaken)
}

func TestSeededUserCanLogInWithTheRightRoleAndID(t *testing.T) {
	svc, _ := newTestService(t)
	ctx := context.Background()
	_, err := svc.SeedDemoUsers(ctx, demoPassword)
	require.NoError(t, err)

	res, err := svc.Login(ctx, "dsa.cst@rub.edu.bt", demoPassword)
	require.NoError(t, err)
	assert.Equal(t, dsaID, res.User.ID)
	assert.Equal(t, []domain.Role{domain.RoleCategoryHead}, res.User.Roles)

	claims, err := svc.tokens.Parse(res.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, dsaID, claims.UserID)
	assert.Equal(t, []domain.Role{domain.RoleCategoryHead}, claims.Roles)

	_, err = svc.Login(ctx, "dsa.cst@rub.edu.bt", "wrong-password")
	assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
}

func TestWithoutTheSeedDemoUsersCannotLogIn(t *testing.T) {
	svc, _ := newTestService(t)
	_, err := svc.Login(context.Background(), "dsa.cst@rub.edu.bt", demoPassword)
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
