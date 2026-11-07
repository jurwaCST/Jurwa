package events

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jurwaCST/Jurwa/pkg/eventbus"
	commonv1 "github.com/jurwaCST/Jurwa/proto/gen/common/v1"
	eventsv1 "github.com/jurwaCST/Jurwa/proto/gen/events/v1"
	"github.com/jurwaCST/Jurwa/services/status-lifecycle/application"
)

type fakeApp struct {
	mu        sync.Mutex
	submitted []string
	routed    []string // "report|position|correlation"
	lapsed    []application.SagaEvent
}

func (f *fakeApp) OnReportSubmitted(_ context.Context, reportID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.submitted = append(f.submitted, reportID)
	return nil
}

func (f *fakeApp) OnReportRouted(_ context.Context, reportID, position, correlation string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.routed = append(f.routed, reportID+"|"+position+"|"+correlation)
	return nil
}

func (f *fakeApp) EscalateOnLapse(_ context.Context, ev application.SagaEvent) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.lapsed = append(f.lapsed, ev)
	return nil
}

func (f *fakeApp) counts() (int, int, int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.submitted), len(f.routed), len(f.lapsed)
}

func TestStart_PassesTheThreeStreamsToTheApp(t *testing.T) {
	bus := eventbus.NewMemory()
	app := &fakeApp{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	Start(ctx, bus, "test", app)

	require.NoError(t, bus.Publish(ctx, eventbus.StreamReportSubmitted, &eventsv1.ReportSubmitted{
		Meta: &eventsv1.EventMeta{EventId: "e0"}, Report: &commonv1.Report{Id: "r1"},
	}))
	require.NoError(t, bus.Publish(ctx, eventbus.StreamReportRouted, &eventsv1.ReportRouted{
		Meta: &eventsv1.EventMeta{EventId: "e1", CorrelationId: "c0"}, ReportId: "r1", TargetPosition: "DSA",
	}))
	require.NoError(t, bus.Publish(ctx, eventbus.StreamSlaLapsed, &eventsv1.SlaLapsed{
		Meta: &eventsv1.EventMeta{EventId: "e2", CorrelationId: "c1"}, ReportId: "r1",
	}))
	assert.Eventually(t, func() bool { s, r, l := app.counts(); return s == 1 && r == 1 && l == 1 }, 2*time.Second, 5*time.Millisecond)

	app.mu.Lock()
	defer app.mu.Unlock()
	assert.Equal(t, []string{"r1"}, app.submitted)
	assert.Equal(t, []string{"r1|DSA|c0"}, app.routed)
	assert.Equal(t, application.SagaEvent{EventID: "e2", CorrelationID: "c1", ReportID: "r1"}, app.lapsed[0])
}

func TestHandlers_BadPayload(t *testing.T) {
	bad := []byte{0xff, 0xff}
	assert.Error(t, OnReportSubmitted(&fakeApp{})(context.Background(), bad))
	assert.Error(t, OnReportRouted(&fakeApp{})(context.Background(), bad))
	assert.Error(t, OnSlaLapsed(&fakeApp{})(context.Background(), bad))
}

// failingSubscriber makes Start log a stopped subscription.
type failingSubscriber struct{ calls chan struct{} }

func (f failingSubscriber) Subscribe(context.Context, string, string, string, eventbus.Handler) error {
	f.calls <- struct{}{}
	return assert.AnError
}

func TestStart_SubscriptionError(t *testing.T) {
	sub := failingSubscriber{calls: make(chan struct{}, 3)}
	Start(context.Background(), sub, "test", &fakeApp{})
	for range 3 {
		select {
		case <-sub.calls:
		case <-time.After(2 * time.Second):
			t.Fatal("Subscribe was not called for all three streams")
		}
	}
}
