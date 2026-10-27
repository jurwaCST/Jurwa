package uuid

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jurwaCST/Jurwa/services/report-submission/domain"
)

func TestNewIDIsUniqueV4UUID(t *testing.T) {
	g := Generator{}
	seen := make(map[string]bool)
	for range 100 {
		id := g.NewID()
		assert.True(t, domain.IsUUID(id), "not a UUID: %s", id)
		assert.Equal(t, byte('4'), id[14], "version must be 4: %s", id)
		assert.False(t, seen[id], "duplicate id %s", id)
		seen[id] = true
	}
}
