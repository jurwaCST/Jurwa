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

// DemoUsers are the accounts for the demo, one per role that cannot register.
// The three Category Head ids are the holders of DSA, DAA and SSO in
// services/department-directory (demoHolderIDs), so a routed report lands on a
// user who can log in. ".example" is a reserved domain, never a real mailbox.
var DemoUsers = []DemoUser{
	{ID: "00000000-0000-4000-8000-000000000001", Email: "dsa.demo@jurwa.example", FullName: "DSA (demo)", Role: domain.RoleCategoryHead},
	{ID: "00000000-0000-4000-8000-000000000002", Email: "daa.demo@jurwa.example", FullName: "DAA (demo)", Role: domain.RoleCategoryHead},
	{ID: "00000000-0000-4000-8000-000000000003", Email: "sso.demo@jurwa.example", FullName: "SSO (demo)", Role: domain.RoleCategoryHead},
	{ID: "00000000-0000-4000-8000-000000000010", Email: "staff.demo@jurwa.example", FullName: "Staff (demo)", Role: domain.RoleStaff},
	{ID: "00000000-0000-4000-8000-000000000020", Email: "admin.demo@jurwa.example", FullName: "System Administrator (demo)", Role: domain.RoleSystemAdmin},
	{ID: "00000000-0000-4000-8000-000000000030", Email: "president.demo@jurwa.example", FullName: "College President (demo)", Role: domain.RolePresident},

	// SCRUM-158: the HoDs receive Academic Block reports and the councillors
	// receive Hostel and Others reports from the SSO. They fix them, so they are
	// Staff, and each is named by position. The ids are the holders in
	// services/department-directory (demoHolderIDs).
	{ID: "00000000-0000-4000-8000-000000000041", Email: "hod.arch@jurwa.example", FullName: "HoD, Architecture", Role: domain.RoleStaff},
	{ID: "00000000-0000-4000-8000-000000000042", Email: "hod.civil@jurwa.example", FullName: "HoD, Civil & Environmental Engineering", Role: domain.RoleStaff},
	{ID: "00000000-0000-4000-8000-000000000043", Email: "hod.ct@jurwa.example", FullName: "HoD, Computing Technologies", Role: domain.RoleStaff},
	{ID: "00000000-0000-4000-8000-000000000044", Email: "hod.eee@jurwa.example", FullName: "HoD, Electrical & Electronics Engineering", Role: domain.RoleStaff},
	{ID: "00000000-0000-4000-8000-000000000045", Email: "hod.mech@jurwa.example", FullName: "HoD, Mechanical Engineering (acting)", Role: domain.RoleStaff},
	{ID: "00000000-0000-4000-8000-000000000046", Email: "hod.sh@jurwa.example", FullName: "HoD, Science & Humanities", Role: domain.RoleStaff},
	{ID: "00000000-0000-4000-8000-000000000051", Email: "councillor.ha@jurwa.example", FullName: "Block Councillor (HA)", Role: domain.RoleStaff},
	{ID: "00000000-0000-4000-8000-000000000052", Email: "councillor.hb@jurwa.example", FullName: "Block Councillor (HB)", Role: domain.RoleStaff},
	{ID: "00000000-0000-4000-8000-000000000053", Email: "councillor.hc@jurwa.example", FullName: "Block Councillor (HC)", Role: domain.RoleStaff},
	{ID: "00000000-0000-4000-8000-000000000054", Email: "councillor.hd@jurwa.example", FullName: "Block Councillor (HD)", Role: domain.RoleStaff},
	{ID: "00000000-0000-4000-8000-000000000055", Email: "councillor.he@jurwa.example", FullName: "Block Councillor (HE)", Role: domain.RoleStaff},
	{ID: "00000000-0000-4000-8000-000000000056", Email: "councillor.hf@jurwa.example", FullName: "Block Councillor (HF)", Role: domain.RoleStaff},
	{ID: "00000000-0000-4000-8000-000000000057", Email: "councillor.nk@jurwa.example", FullName: "Block Councillor (NK)", Role: domain.RoleStaff},
	{ID: "00000000-0000-4000-8000-000000000058", Email: "councillor.rka@jurwa.example", FullName: "Block Councillor (RKA)", Role: domain.RoleStaff},
	{ID: "00000000-0000-4000-8000-000000000059", Email: "councillor.rkb@jurwa.example", FullName: "Block Councillor (RKB)", Role: domain.RoleStaff},
	{ID: "00000000-0000-4000-8000-000000000060", Email: "councillor.games@jurwa.example", FullName: "Games and Health Councillor", Role: domain.RoleStaff},
	{ID: "00000000-0000-4000-8000-000000000061", Email: "councillor.mess@jurwa.example", FullName: "Deputy Chief Councillor (Mess)", Role: domain.RoleStaff},
	{ID: "00000000-0000-4000-8000-000000000062", Email: "councillor.electrical@jurwa.example", FullName: "Electrical Maintenance Councillor", Role: domain.RoleStaff},
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
