package events

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
	"github.com/jurwaCST/Jurwa/services/routing-engine/adapters/out/eventpublisher"
	"github.com/jurwaCST/Jurwa/services/routing-engine/application"
	"github.com/jurwaCST/Jurwa/services/routing-engine/domain"
)

// fakeRouter records what the consumer asks for.
type fakeRouter struct {
	src string
	req domain.Request
	err error
}

func (f *fakeRouter) RouteSubmitted(_ context.Context, src string, req domain.Request) (domain.Decision, error) {
	f.src, f.req = src, req
	return domain.Decision{}, f.err
}

func submitted(t *testing.T, eventID string, c commonv1.Category) []byte {
	t.Helper()
	b, err := proto.Marshal(&eventsv1.ReportSubmitted{
		Meta: &eventsv1.EventMeta{EventId: eventID, Producer: "report-submission", CorrelationId: "r1"},
		Report: &commonv1.Report{
			Id: "r1", ReporterId: "u1", Category: c, Description: "send this to DAA",
			Location: &commonv1.Location{Building: "Hostel B", Wing: "East", Area: "Bathroom 2"},
		},
	})
	require.NoError(t, err)
	return b
}

func TestHandleRoutesTheSubmittedReport(t *testing.T) {
	r := &fakeRouter{}
	require.NoError(t, New(r).Handle(context.Background(), submitted(t, "ev-1", commonv1.Category_HOSTEL)))
	assert.Equal(t, "ev-1", r.src)
	assert.Equal(t, domain.Request{ReportID: "r1", Category: domain.Hostel, Building: "Hostel B"}, r.req,
		"only the report id, category and building: never the description")
}

func TestHandleMapsEveryCategory(t *testing.T) {
	for c, want := range map[commonv1.Category]domain.Category{
		commonv1.Category_HOSTEL:               domain.Hostel,
		commonv1.Category_ACADEMIC_BLOCK:       domain.AcademicBlock,
		commonv1.Category_OTHERS:               domain.Others,
		commonv1.Category_CATEGORY_UNSPECIFIED: "",
	} {
		r := &fakeRouter{}
		_ = New(r).Handle(context.Background(), submitted(t, "ev-1", c))
		assert.Equal(t, want, r.req.Category, c.String())
	}
}

// An error makes the bus try again, then move the message to the dead-letter stream.
func TestHandleReturnsErrors(t *testing.T) {
	assert.Error(t, New(&fakeRouter{}).Handle(context.Background(), []byte("not protobuf")), "bad payload")

	r := &fakeRouter{err: domain.ErrUnavailable}
	assert.ErrorIs(t, New(r).Handle(context.Background(), submitted(t, "ev-1", commonv1.Category_HOSTEL)), domain.ErrUnavailable)
}

// ---- the whole flow on the in-memory bus ----

type rules struct{}

func (rules) GetRule(_ context.Context, c domain.Category, _ string) (domain.Rule, error) {
	pos := map[domain.Category]string{domain.Hostel: "DSA", domain.AcademicBlock: "DAA", domain.Others: "SSO"}[c]
	if pos == "" {
		return domain.Rule{}, domain.ErrRuleNotFound
	}
	return domain.Rule{ID: "rule-" + pos, TargetPosition: pos}, nil
}

type holders struct{}

func (holders) GetHolder(_ context.Context, pos string) (string, error) { return "user-" + pos, nil }

// report.submitted in, report.routed out, with the Category Head from the
// directory; a redelivered report.submitted gives the same report.routed id.
func TestSubmittedReportIsRoutedOnTheBus(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	bus := eventbus.NewMemory()
	n := 0
	pub := eventpublisher.New(bus, func() string { n++; return "id-" + string(rune('0'+n)) }, time.Now)
	app := application.NewService(rules{}, holders{}, pub)
	go func() { _ = bus.Subscribe(ctx, eventbus.StreamReportSubmitted, "routing-engine", "c1", New(app).Handle) }()

	got := make(chan *eventsv1.ReportRouted, 4)
	go func() {
		_ = bus.Subscribe(ctx, eventbus.StreamReportRouted, "report-history", "c1", func(_ context.Context, p []byte) error {
			ev := &eventsv1.ReportRouted{}
			if err := proto.Unmarshal(p, ev); err != nil {
				return err
			}
			got <- ev
			return nil
		})
	}()

	sub := &eventsv1.ReportSubmitted{}
	require.NoError(t, proto.Unmarshal(submitted(t, "ev-77", commonv1.Category_ACADEMIC_BLOCK), sub))
	require.NoError(t, bus.Publish(ctx, eventbus.StreamReportSubmitted, sub))
	require.NoError(t, bus.Publish(ctx, eventbus.StreamReportSubmitted, sub)) // delivered twice

	var evs []*eventsv1.ReportRouted
	for range 2 {
		select {
		case ev := <-got:
			evs = append(evs, ev)
		case <-time.After(2 * time.Second):
			t.Fatal("no report.routed")
		}
	}
	assert.Equal(t, "r1", evs[0].GetReportId())
	assert.Equal(t, commonv1.Category_ACADEMIC_BLOCK, evs[0].GetCategory())
	assert.Equal(t, "DAA", evs[0].GetTargetPosition())
	assert.Equal(t, "user-DAA", evs[0].GetCategoryHeadId())
	assert.Equal(t, evs[0].GetMeta().GetEventId(), evs[1].GetMeta().GetEventId(), "readers can drop the copy")
}
