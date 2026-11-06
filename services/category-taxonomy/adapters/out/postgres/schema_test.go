package postgres

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The adapter embeds its own copy of the table design. It must stay the same
// as docs/design/schema/category-taxonomy.sql.
func TestSchemaMatchesTheDesignFile(t *testing.T) {
	design, err := os.ReadFile("../../../../../docs/design/schema/category-taxonomy.sql")
	require.NoError(t, err)
	assert.Equal(t, string(design), schemaSQL)
}
