package mongo

import (
	"time"

	"github.com/jurwaCST/Jurwa/services/report-submission/domain"
)

// document is a report as stored in MongoDB (field names from the schema).
type document struct {
	ID          string    `bson:"_id"`
	ReporterID  string    `bson:"reporter_id"`
	Category    string    `bson:"category"`
	Location    location  `bson:"location"`
	Description string    `bson:"description"`
	PhotoID     *string   `bson:"photo_id,omitempty"`
	Status      string    `bson:"status"`
	CreatedAt   time.Time `bson:"created_at"`
}

type location struct {
	Building string `bson:"building"`
	Wing     string `bson:"wing,omitempty"`
	Area     string `bson:"area,omitempty"`
}

func toDocument(r domain.Report) document {
	return document{
		ID:          r.ID,
		ReporterID:  r.ReporterID,
		Category:    string(r.Category),
		Location:    location{Building: r.Location.Building, Wing: r.Location.Wing, Area: r.Location.Area},
		Description: r.Description,
		PhotoID:     r.PhotoID,
		Status:      string(r.Status),
		CreatedAt:   r.CreatedAt,
	}
}

func (d document) toReport() domain.Report {
	return domain.Report{
		ID:          d.ID,
		ReporterID:  d.ReporterID,
		Category:    domain.Category(d.Category),
		Location:    domain.Location{Building: d.Location.Building, Wing: d.Location.Wing, Area: d.Location.Area},
		Description: d.Description,
		PhotoID:     d.PhotoID,
		Status:      domain.Status(d.Status),
		CreatedAt:   d.CreatedAt.UTC(),
	}
}
