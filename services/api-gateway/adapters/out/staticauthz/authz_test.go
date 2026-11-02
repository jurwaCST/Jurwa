package staticauthz

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jurwaCST/Jurwa/services/api-gateway/domain"
)

func TestCheckPermission(t *testing.T) {
	tests := []struct {
		role, action string
		want         bool
	}{
		{domain.RoleStudent, domain.ActionReportSubmit, true},
		{domain.RoleCategoryHead, domain.ActionReportSubmit, false},
		{domain.RoleCategoryHead, domain.ActionReportRead, true},
		{"PRESIDENT", domain.ActionReportRead, false},
		{"PRESIDENT", domain.ActionCategoryList, true},
		{domain.RoleStudent, "unknown.action", false},
	}
	for _, tt := range tests {
		t.Run(tt.role+" "+tt.action, func(t *testing.T) {
			got, err := New().CheckPermission(context.Background(), tt.role, tt.action, "x")
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
