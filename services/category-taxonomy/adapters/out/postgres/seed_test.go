package postgres

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jurwaCST/Jurwa/services/category-taxonomy/domain"
)

const oldOthersDescription = "Campus-wide issues such as Wi-Fi, roads and grounds"

// SCRUM-150: the database seed uses the same descriptions as the code.
func TestSeedHasTheDefaultDescriptions(t *testing.T) {
	for _, c := range domain.DefaultCategories() {
		assert.Contains(t, schemaSQL, "'"+c.Description+"'", "seed row for %s", c.Category)
	}
}

// A database created before SCRUM-150 still has the old Others text, and the
// seed's INSERT ... ON CONFLICT DO NOTHING does not change it. An UPDATE
// replaces only that old default text, so an administrator's own wording is kept.
func TestSchemaReplacesOnlyTheOldOthersText(t *testing.T) {
	others, _ := domain.Find(domain.DefaultCategories(), domain.Others)
	update := "UPDATE categories SET description = '" + others.Description + "'"

	assert.Contains(t, schemaSQL, update)
	assert.Contains(t, schemaSQL, "WHERE code = 'OTHERS' AND description = '"+oldOthersDescription+"'")
	assert.Less(t, strings.Index(schemaSQL, "INSERT INTO categories"), strings.Index(schemaSQL, update), "the UPDATE runs after the seed")
}
