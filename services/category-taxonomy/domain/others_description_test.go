package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// SCRUM-150: Others are places, not services (docs/design/real-locations.md),
// so its description names places and no longer mentions Wi-Fi or roads.
func TestOthersDescriptionNamesPlaces(t *testing.T) {
	others, err := Find(DefaultCategories(), Others)
	require.NoError(t, err)

	assert.Equal(t, "Campus places such as the library, mess, gym and sports grounds", others.Description)
	assert.NotContains(t, others.Description, "Wi-Fi")
}
