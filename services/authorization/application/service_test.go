package application

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jurwaCST/Jurwa/services/authorization/adapters/out/memory"
	"github.com/jurwaCST/Jurwa/services/authorization/domain"
)

// failingRepo is a fake repository whose Load always fails.
type failingRepo struct{}

func (failingRepo) Load(context.Context) (domain.Policy, error) {
	return domain.Policy{}, errors.New("store down")
}

func TestCheckPermission(t *testing.T) {
	svc := NewService(memory.New())
	ctx := context.Background()

	tests := []struct {
		name    string
		role    domain.Role
		action  string
		want    bool
		wantErr error
	}{
		{"student submits a report", domain.Student, domain.ActionReportSubmit, true, nil},
		{"president changes status is denied", domain.President, domain.ActionReportStatusChange, false, nil},
		{"unknown action is denied", domain.SystemAdmin, "report.delete", false, nil},
		{"role unspecified is an error", domain.RoleUnspecified, domain.ActionMeRead, false, domain.ErrRoleRequired},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := svc.CheckPermission(ctx, tt.role, tt.action, "report")
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, d.Allowed)
		})
	}
}

func TestCheckPermissionStoreError(t *testing.T) {
	_, err := NewService(failingRepo{}).CheckPermission(context.Background(), domain.Student, domain.ActionMeRead, "me")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "store down")
}
