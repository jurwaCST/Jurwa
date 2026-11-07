package events

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/jurwaCST/Jurwa/pkg/eventbus"
	commonv1 "github.com/jurwaCST/Jurwa/proto/gen/common/v1"
	eventsv1 "github.com/jurwaCST/Jurwa/proto/gen/events/v1"
	"github.com/jurwaCST/Jurwa/services/sla-tracking/domain"
)

type fakeTimers struct {
	mu       sync.Mutex
	started  []string // "report|category"
	stopped  []string
	startErr error
	stopErr  error
}

func (f *fakeTimers) StartTimer(_ context.Context, reportID string, c domain.Category) (domain.Timer, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.started = append(f.started, reportID+"|"+c.String())
	return domain.Timer{}, f.startErr
}

func (f *fakeTimers) StopTimer(_ context.Context, reportID string) (domain.Timer, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.stopped = append(f.stopped, reportID)
	return domain.Timer{}, f.stopErr
}

func (f *fakeTimers) counts() (int, int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.started), len(f.stopped)
}

func TestStart_RoutedStartsAndResolvedStops(t *testing.T) {
	bus := eventbus.NewMemory()
	app := &fakeTimers{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	Start(ctx, bus, "test", app)

	require.NoError(t, bus.Publish(ctx, eventbus.StreamReportRouted, &eventsv1.ReportRouted{
		Meta: &eventsv1.EventMeta{EventId: "e1"}, ReportId: "r1", Category: commonv1.Category_ACADEMIC_BLOCK,
	}))
	// Only RESOLVED stops the timer.
	require.NoError(t, bus.Publish(ctx, eventbus.StreamStatusChanged, &eventsv1.StatusChanged{
		Meta: &eventsv1.EventMeta{EventId: "e2"}, ReportId: "r1", ToStatus: commonv1.Status_IN_PROGRESS,
	}))
	require.NoError(t, bus.Publish(ctx, eventbus.StreamStatusChanged, &eventsv1.StatusChanged{
		Meta: &eventsv1.EventMeta{EventId: "e3"}, ReportId: "r1", ToStatus: commonv1.Status_RESOLVED,
	}))

	assert.Eventually(t, func() bool { s, p := app.counts(); return s == 1 && p == 1 }, 2*time.Second, 5*time.Millisecond)
	time.Sleep(20 * time.Millisecond)
	app.mu.Lock()
	defer app.mu.Unlock()
	assert.Equal(t, []string{"r1|ACADEMIC_BLOCK"}, app.started)
	assert.Equal(t, []string{"r1"}, app.stopped, "IN_PROGRESS did not stop it")
}

func payload(t *testing.T, m proto.Message) []byte {
	b, err := proto.Marshal(m)
	require.NoError(t, err)
	return b
}

func TestOnReportRouted_Errors(t *testing.T) {
	ctx := context.Background()
	p := payload(t, &eventsv1.ReportRouted{ReportId: "r1", Category: commonv1.Category_HOSTEL})

	assert.NoError(t, OnReportRouted(&fakeTimers{startErr: domain.ErrAlreadyExists})(ctx, p), "a repeated event counts as done")
	noWindow := errors.New("no SLA window set for this category")
	assert.ErrorIs(t, OnReportRouted(&fakeTimers{startErr: noWindow})(ctx, p), noWindow, "returned, so it ends in the dead stream")
	assert.Error(t, OnReportRouted(&fakeTimers{})(ctx, []byte{0xff, 0xff}))
}

func TestOnStatusChanged_Errors(t *testing.T) {
	ctx := context.Background()
	p := payload(t, &eventsv1.StatusChanged{ReportId: "r1", ToStatus: commonv1.Status_RESOLVED})

	assert.NoError(t, OnStatusChanged(&fakeTimers{stopErr: domain.ErrNotFound})(ctx, p), "a report without a timer is ignored")
	down := errors.New("store down")
	assert.ErrorIs(t, OnStatusChanged(&fakeTimers{stopErr: down})(ctx, p), down)
	assert.Error(t, OnStatusChanged(&fakeTimers{})(ctx, []byte{0xff, 0xff}))
}

type failingSubscriber struct{ calls chan struct{} }

func (f failingSubscriber) Subscribe(context.Context, string, string, string, eventbus.Handler) error {
	f.calls <- struct{}{}
	return assert.AnError
}

func TestStart_SubscriptionError(t *testing.T) {
	sub := failingSubscriber{calls: make(chan struct{}, 2)}
	Start(context.Background(), sub, "test", &fakeTimers{})
	for range 2 {
		select {
		case <-sub.calls:
		case <-time.After(2 * time.Second):
			t.Fatal("Subscribe was not called for both streams")
		}
	}
}
