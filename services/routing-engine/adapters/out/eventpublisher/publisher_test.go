package eventpublisher

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/jurwaCST/Jurwa/pkg/eventbus"
	commonv1 "github.com/jurwaCST/Jurwa/proto/gen/common/v1"
	eventsv1 "github.com/jurwaCST/Jurwa/proto/gen/events/v1"
	"github.com/jurwaCST/Jurwa/services/routing-engine/domain"
	"github.com/jurwaCST/Jurwa/services/routing-engine/ports"
)

var fixedNow = time.Date(2026, 11, 7, 10, 0, 0, 0, time.UTC)

// read returns the routed events on the bus, in order.
func read(t *testing.T, bus *eventbus.Memory, n int) []*eventsv1.ReportRouted {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	got := make(chan *eventsv1.ReportRouted, n)
	go func() {
		_ = bus.Subscribe(ctx, eventbus.StreamReportRouted, "test", "c1", func(_ context.Context, p []byte) error {
			ev := &eventsv1.ReportRouted{}
			require.NoError(t, proto.Unmarshal(p, ev))
			got <- ev
			return nil
		})
	}()
	defer cancel()
	out := make([]*eventsv1.ReportRouted, 0, n)
	for range n {
		select {
		case ev := <-got:
			out = append(out, ev)
		case <-time.After(2 * time.Second):
			t.Fatalf("got %d of %d events", len(out), n)
		}
	}
	return out
}

func routed(src string) ports.ReportRouted {
	return ports.ReportRouted{
		ReportID: "r1", Category: domain.Hostel, CategoryHeadID: "user-dsa",
		TargetPosition: "DSA", RuleID: "rule-hostel", SourceEventID: src,
	}
}

func TestPublishReportRoutedSendsEveryField(t *testing.T) {
	bus := eventbus.NewMemory()
	p := New(bus, func() string { return "new-id" }, func() time.Time { return fixedNow })
	require.NoError(t, p.PublishReportRouted(context.Background(), routed("")))

	ev := read(t, bus, 1)[0]
	assert.Equal(t, "r1", ev.GetReportId())
	assert.Equal(t, commonv1.Category_HOSTEL, ev.GetCategory())
	assert.Equal(t, "user-dsa", ev.GetCategoryHeadId())
	assert.Equal(t, "DSA", ev.GetTargetPosition())
	assert.Equal(t, "rule-hostel", ev.GetRuleId())
	assert.Equal(t, "new-id", ev.GetMeta().GetEventId(), "a routing from the RPC gets a new id")
	assert.Equal(t, "routing-engine", ev.GetMeta().GetProducer())
	assert.Equal(t, "r1", ev.GetMeta().GetCorrelationId(), "the report id links the flow, like report-submission")
	assert.True(t, fixedNow.Equal(ev.GetMeta().GetOccurredAt().AsTime()))
}

// report.submitted can arrive twice. Routing it twice must give the same
// report.routed event id, so readers drop the copy (pkg/eventbus rule).
func TestTheSameSubmittedEventGivesTheSameRoutedEventID(t *testing.T) {
	bus := eventbus.NewMemory()
	p := New(bus, func() string { return "random" }, func() time.Time { return fixedNow })
	ctx := context.Background()
	require.NoError(t, p.PublishReportRouted(ctx, routed("ev-submitted-1")))
	require.NoError(t, p.PublishReportRouted(ctx, routed("ev-submitted-1")))
	require.NoError(t, p.PublishReportRouted(ctx, routed("ev-submitted-2")))

	evs := read(t, bus, 3)
	first := evs[0].GetMeta().GetEventId()
	assert.Equal(t, first, evs[1].GetMeta().GetEventId(), "redelivery: same id")
	assert.NotEqual(t, first, evs[2].GetMeta().GetEventId(), "another report.submitted: another id")
	assert.NotEqual(t, "random", first)
	assert.Regexp(t, `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`, first, "still a UUID")
}

func TestCategoriesMapToTheProtoEnum(t *testing.T) {
	for c, want := range map[domain.Category]commonv1.Category{
		domain.Hostel:        commonv1.Category_HOSTEL,
		domain.AcademicBlock: commonv1.Category_ACADEMIC_BLOCK,
		domain.Others:        commonv1.Category_OTHERS,
		"LIBRARY":            commonv1.Category_CATEGORY_UNSPECIFIED,
	} {
		assert.Equal(t, want, category(c), string(c))
	}
}
