package mongo

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/jurwaCST/Jurwa/services/report-submission/domain"
)

// The document shape is fixed in docs/design/database-schema.md section 5.
func TestDocumentRoundTrip(t *testing.T) {
	photo := "0b9c8d7e-6f5a-4b3c-9d2e-1f0a9b8c7d6e"
	r := domain.Report{
		ID:          "6f1c2a7e-3b4d-4c5e-8f90-1a2b3c4d5e6f",
		ReporterID:  "11111111-2222-4333-8444-555555555555",
		Category:    domain.CategoryOthers,
		Location:    domain.Location{Building: "Library", Wing: "North", Area: "Reading room"},
		Description: "Fan not working",
		PhotoID:     &photo,
		Status:      domain.StatusOpen,
		CreatedAt:   time.Date(2026, 11, 6, 4, 30, 0, 0, time.UTC),
	}

	doc := toDocument(r)
	assert.Equal(t, r.ID, doc.ID)
	assert.Equal(t, "OTHERS", doc.Category)
	assert.Equal(t, "Library", doc.Location.Building)
	assert.Equal(t, &photo, doc.PhotoID)
	assert.Equal(t, r, doc.toReport())
}

func TestDocumentWithoutPhoto(t *testing.T) {
	r := domain.Report{ID: "x", Status: domain.StatusOpen, CreatedAt: time.Unix(0, 0).UTC()}
	doc := toDocument(r)
	assert.Nil(t, doc.PhotoID)
	assert.Nil(t, doc.toReport().PhotoID)
}
