package main

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/stretchr/testify/assert"

	"github.com/jurwaCST/Jurwa/services/reassignment/adapters/out/logpublisher"
	"github.com/jurwaCST/Jurwa/services/reassignment/adapters/out/memory"
	"github.com/jurwaCST/Jurwa/services/reassignment/application"
	"github.com/jurwaCST/Jurwa/services/reassignment/domain"
)

func TestEnvOr(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		fallback string
		want     string
	}{
		{name: "uses value when set", value: "9100", fallback: "9000", want: "9100"},
		{name: "uses fallback when empty", value: "", fallback: "9000", want: "9000"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("JURWA_TEST_VAR", tt.value)
			assert.Equal(t, tt.want, envOr("JURWA_TEST_VAR", tt.fallback))
		})
	}
}

func TestSeedAssignments(t *testing.T) {
	app := application.NewService(memory.New(), logpublisher.New(slog.Default()), time.Now, uuid.NewString)
	seedAssignments(app, " 11111111-1111-1111-1111-111111111111:22222222-2222-2222-2222-222222222222:DSA, not-a-uuid:x,,"+
		"33333333-3333-3333-3333-333333333333:22222222-2222-2222-2222-222222222222")

	for id, pos := range map[string]string{
		"11111111-1111-1111-1111-111111111111": "DSA",
		"33333333-3333-3333-3333-333333333333": "",
	} {
		list, err := app.ListReassignments(context.Background(), domain.RoleSystemAdmin, id)
		assert.NoError(t, err)
		if assert.Len(t, list, 1, id) {
			assert.Equal(t, pos, list[0].AssigneePositionCode)
		}
	}
}
