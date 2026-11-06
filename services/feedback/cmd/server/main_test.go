package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestOpenRepoWithoutDatabaseURLUsesMemory(t *testing.T) {
	repo, closeRepo, err := openRepo("")
	assert.NoError(t, err)
	assert.NotNil(t, repo)
	closeRepo()
}

// A wrong DATABASE_URL stops the service instead of silently using memory.
func TestOpenRepoWithABadURLFails(t *testing.T) {
	_, _, err := openRepo("not a database url")
	assert.Error(t, err)
}
