package application

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jurwaCST/Jurwa/services/identity/domain"
)

// Errors from SeedDemoUsers.
var (
	ErrDemoPasswordMissing = errors.New("demo password is required")
	ErrDemoEmailTaken      = errors.New("a demo email is already used by another account")
)

// DemoUser is one account created by SeedDemoUsers. These are not real people.
type DemoUser struct {
	ID       string
	Email    string
	FullName string
	Role     domain.Role
}

// DemoUsers are the accounts for the demo. SCRUM-159: they use the CST email
// format and new ids (...01NN), so a database that still holds the earlier
// @jurwa.example accounts (ids ...00NN) starts without a clash; those old
// accounts are simply no longer used.
// The three Category Head ids are the holders of DSA, DAA and SSO in
// services/department-directory (demoHolderIDs), so a routed report lands on a
// user who can log in. Jurwa sends no email, so these addresses never receive mail.
var DemoUsers = []DemoUser{
	{ID: "00000000-0000-4000-8000-000000000101", Email: "dsa.cst@rub.edu.bt", FullName: "Dean of Student Affairs", Role: domain.RoleCategoryHead},
	{ID: "00000000-0000-4000-8000-000000000102", Email: "daa.cst@rub.edu.bt", FullName: "Dean of Academic Affairs", Role: domain.RoleCategoryHead},
	{ID: "00000000-0000-4000-8000-000000000103", Email: "sso.cst@rub.edu.bt", FullName: "Student Service Officer", Role: domain.RoleCategoryHead},
	{ID: "00000000-0000-4000-8000-000000000110", Email: "staff.cst@rub.edu.bt", FullName: "Student Service Officer (male)", Role: domain.RoleStaff},
	{ID: "00000000-0000-4000-8000-000000000120", Email: "admin.cst@rub.edu.bt", FullName: "System Administrator", Role: domain.RoleSystemAdmin},
	{ID: "00000000-0000-4000-8000-000000000130", Email: "president.cst@rub.edu.bt", FullName: "College President", Role: domain.RolePresident},

	// SCRUM-158: the HoDs receive Academic Block reports and the councillors
	// receive Hostel and Others reports from the SSO. They fix them, so they are
	// Staff, and each is named by position. The ids are the holders in
	// services/department-directory (demoHolderIDs).
	{ID: "00000000-0000-4000-8000-000000000141", Email: "hod.arch.cst@rub.edu.bt", FullName: "HoD, Architecture", Role: domain.RoleStaff},
	{ID: "00000000-0000-4000-8000-000000000142", Email: "hod.civil.cst@rub.edu.bt", FullName: "HoD, Civil & Environmental Engineering", Role: domain.RoleStaff},
	{ID: "00000000-0000-4000-8000-000000000143", Email: "hod.ct.cst@rub.edu.bt", FullName: "HoD, Computing Technologies", Role: domain.RoleStaff},
	{ID: "00000000-0000-4000-8000-000000000144", Email: "hod.eee.cst@rub.edu.bt", FullName: "HoD, Electrical & Electronics Engineering", Role: domain.RoleStaff},
	{ID: "00000000-0000-4000-8000-000000000145", Email: "hod.mech.cst@rub.edu.bt", FullName: "HoD, Mechanical Engineering (acting)", Role: domain.RoleStaff},
	{ID: "00000000-0000-4000-8000-000000000146", Email: "hod.sh.cst@rub.edu.bt", FullName: "HoD, Science & Humanities", Role: domain.RoleStaff},
	{ID: "00000000-0000-4000-8000-000000000151", Email: "ha.cst@rub.edu.bt", FullName: "Block Councillor (HA)", Role: domain.RoleStaff},
	{ID: "00000000-0000-4000-8000-000000000152", Email: "hb.cst@rub.edu.bt", FullName: "Block Councillor (HB)", Role: domain.RoleStaff},
	{ID: "00000000-0000-4000-8000-000000000153", Email: "hc.cst@rub.edu.bt", FullName: "Block Councillor (HC)", Role: domain.RoleStaff},
	{ID: "00000000-0000-4000-8000-000000000154", Email: "hd.cst@rub.edu.bt", FullName: "Block Councillor (HD)", Role: domain.RoleStaff},
	{ID: "00000000-0000-4000-8000-000000000155", Email: "he.cst@rub.edu.bt", FullName: "Block Councillor (HE)", Role: domain.RoleStaff},
	{ID: "00000000-0000-4000-8000-000000000156", Email: "hf.cst@rub.edu.bt", FullName: "Block Councillor (HF)", Role: domain.RoleStaff},
	{ID: "00000000-0000-4000-8000-000000000157", Email: "nk.cst@rub.edu.bt", FullName: "Block Councillor (NK)", Role: domain.RoleStaff},
	{ID: "00000000-0000-4000-8000-000000000158", Email: "rka.cst@rub.edu.bt", FullName: "Block Councillor (RKA)", Role: domain.RoleStaff},
	{ID: "00000000-0000-4000-8000-000000000159", Email: "rkb.cst@rub.edu.bt", FullName: "Block Councillor (RKB)", Role: domain.RoleStaff},
	{ID: "00000000-0000-4000-8000-000000000160", Email: "games.cst@rub.edu.bt", FullName: "Games and Health Councillor", Role: domain.RoleStaff},
	{ID: "00000000-0000-4000-8000-000000000161", Email: "mess.cst@rub.edu.bt", FullName: "Deputy Chief Councillor (Mess)", Role: domain.RoleStaff},
	{ID: "00000000-0000-4000-8000-000000000162", Email: "electrical.cst@rub.edu.bt", FullName: "Electrical Maintenance Councillor", Role: domain.RoleStaff},

	// SCRUM-159: a demo student in the student number format, so the demo does
	// not have to register one. Not a real student.
	{ID: "00000000-0000-4000-8000-000000000199", Email: "02240300.cst@rub.edu.bt", FullName: "Demo Student", Role: domain.RoleStudent},
}

// SeedDemoUsers creates the DemoUsers that do not exist yet, all with the same
// password, and returns how many it created. It is safe to run on every start:
// an account that is already there is left as it is. It goes through the
// UserRepository port, so it works with the memory and the PostgreSQL adapter.
// The password is checked first, so a bad password creates nothing.
func (s *Service) SeedDemoUsers(ctx context.Context, password string) (int, error) {
	if strings.TrimSpace(password) == "" {
		return 0, ErrDemoPasswordMissing
	}
	if err := domain.ValidatePassword(password); err != nil {
		return 0, err
	}

	created := 0
	for _, d := range DemoUsers {
		existing, err := s.users.FindByEmail(ctx, d.Email)
		switch {
		case err == nil && existing.ID == d.ID:
			continue // seeded on an earlier start
		case err == nil:
			return created, fmt.Errorf("%w: %s", ErrDemoEmailTaken, d.Email)
		case !errors.Is(err, domain.ErrUserNotFound):
			return created, err
		}

		u, err := domain.NewUserWithRole(d.ID, d.Email, d.FullName, password, d.Role, s.now().UTC())
		if err != nil {
			return created, fmt.Errorf("demo user %s: %w", d.Email, err)
		}
		if err := s.users.Create(ctx, u); err != nil {
			if errors.Is(err, domain.ErrEmailTaken) {
				continue // another replica seeded it at the same moment
			}
			return created, fmt.Errorf("demo user %s: %w", d.Email, err)
		}
		created++
	}
	return created, nil
}
