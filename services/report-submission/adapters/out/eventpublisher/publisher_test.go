package eventpublisher

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/jurwaCST/Jurwa/pkg/eventbus"
	commonv1 "github.com/jurwaCST/Jurwa/proto/gen/common/v1"
	eventsv1 "github.com/jurwaCST/Jurwa/proto/gen/events/v1"
	"github.com/jurwaCST/Jurwa/services/report-submission/domain"
)

var created = time.Date(2026, 11, 6, 4, 30, 0, 0, time.UTC)

func sampleReport() domain.Report {
	photo := "0b9c8d7e-6f5a-4b3c-9d2e-1f0a9b8c7d6e"
	return domain.Report{
		ID:          "6f1c2a7e-3b4d-4c5e-8f90-1a2b3c4d5e6f",
		ReporterID:  "11111111-2222-4333-8444-555555555555",
		Category:    domain.CategoryHostel,
		Location:    domain.Location{Building: "Hostel B", Wing: "East", Area: "Bathroom 2"},
		Description: "Tap is leaking",
		PhotoID:     &photo,
		Status:      domain.StatusOpen,
		CreatedAt:   created,
	}
}

func TestReportSubmittedGoesOnTheStream(t *testing.T) {
	bus := eventbus.NewMemory()
	now := time.Date(2026, 11, 6, 5, 0, 0, 0, time.UTC)
	pub := New(bus, func() string { return "evt-1" }, func() time.Time { return now })

	require.NoError(t, pub.ReportSubmitted(context.Background(), sampleReport()))

	got := make(chan []byte, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		_ = bus.Subscribe(ctx, eventbus.StreamReportSubmitted, "test", "c1", func(_ context.Context, p []byte) error {
			got <- p
			return nil
		})
	}()

	var ev eventsv1.ReportSubmitted
	select {
	case p := <-got:
		require.NoError(t, proto.Unmarshal(p, &ev))
	case <-time.After(5 * time.Second):
		t.Fatal("no event on events:report.submitted")
	}

	assert.Equal(t, "evt-1", ev.GetMeta().GetEventId())
	assert.Equal(t, "report-submission", ev.GetMeta().GetProducer())
	assert.Equal(t, "6f1c2a7e-3b4d-4c5e-8f90-1a2b3c4d5e6f", ev.GetMeta().GetCorrelationId())
	assert.True(t, now.Equal(ev.GetMeta().GetOccurredAt().AsTime()))

	r := ev.GetReport()
	assert.Equal(t, "6f1c2a7e-3b4d-4c5e-8f90-1a2b3c4d5e6f", r.GetId())
	assert.Equal(t, "11111111-2222-4333-8444-555555555555", r.GetReporterId())
	assert.Equal(t, commonv1.Category_HOSTEL, r.GetCategory())
	assert.Equal(t, "Hostel B", r.GetLocation().GetBuilding())
	assert.Equal(t, "East", r.GetLocation().GetWing())
	assert.Equal(t, "Bathroom 2", r.GetLocation().GetArea())
	assert.Equal(t, "Tap is leaking", r.GetDescription())
	assert.Equal(t, "0b9c8d7e-6f5a-4b3c-9d2e-1f0a9b8c7d6e", r.GetPhotoId())
	assert.Equal(t, commonv1.Status_OPEN, r.GetStatus())
	assert.True(t, created.Equal(r.GetCreatedAt().AsTime()))
}

func TestCategoriesAreMapped(t *testing.T) {
	cases := map[domain.Category]commonv1.Category{
		domain.CategoryHostel:        commonv1.Category_HOSTEL,
		domain.CategoryAcademicBlock: commonv1.Category_ACADEMIC_BLOCK,
		domain.CategoryOthers:        commonv1.Category_OTHERS,
		domain.Category("??"):        commonv1.Category_CATEGORY_UNSPECIFIED,
	}
	for in, want := range cases {
		r := sampleReport()
		r.Category = in
		assert.Equal(t, want, toProto(r).GetCategory(), in)
	}
}

func TestReportWithoutPhotoHasNoPhotoID(t *testing.T) {
	r := sampleReport()
	r.PhotoID = nil
	assert.Nil(t, toProto(r).PhotoId)
}

type failingBus struct{}

func (failingBus) Publish(context.Context, string, proto.Message) error {
	return errors.New("redis down")
}

func TestPublishErrorIsReturned(t *testing.T) {
	pub := New(failingBus{}, func() string { return "evt-1" }, time.Now)
	assert.Error(t, pub.ReportSubmitted(context.Background(), sampleReport()))
}
