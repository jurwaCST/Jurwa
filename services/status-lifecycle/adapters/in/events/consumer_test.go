package events

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jurwaCST/Jurwa/pkg/eventbus"
	eventsv1 "github.com/jurwaCST/Jurwa/proto/gen/events/v1"
	"github.com/jurwaCST/Jurwa/services/status-lifecycle/application"
)

type fakeSaga struct {
	mu  sync.Mutex
	got []application.SagaEvent
}

func (f *fakeSaga) EscalateOnLapse(_ context.Context, ev application.SagaEvent) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.got = append(f.got, ev)
	return nil
}

func (f *fakeSaga) calls() []application.SagaEvent {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]application.SagaEvent(nil), f.got...)
}

func TestStart_PassesSlaLapsedToTheApp(t *testing.T) {
	bus := eventbus.NewMemory()
	app := &fakeSaga{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	Start(ctx, bus, "test", app)

	require.NoError(t, bus.Publish(ctx, eventbus.StreamSlaLapsed, &eventsv1.SlaLapsed{
		Meta: &eventsv1.EventMeta{EventId: "e1", CorrelationId: "c1"}, ReportId: "r1",
	}))
	assert.Eventually(t, func() bool { return len(app.calls()) == 1 }, 2*time.Second, 5*time.Millisecond)
	assert.Equal(t, application.SagaEvent{EventID: "e1", CorrelationID: "c1", ReportID: "r1"}, app.calls()[0])
}

func TestOnSlaLapsed_BadPayload(t *testing.T) {
	err := OnSlaLapsed(&fakeSaga{})(context.Background(), []byte{0xff, 0xff})
	assert.Error(t, err)
}

// failingSubscriber makes Start log a stopped subscription.
type failingSubscriber struct{ called chan struct{} }

func (f failingSubscriber) Subscribe(context.Context, string, string, string, eventbus.Handler) error {
	close(f.called)
	return assert.AnError
}

func TestStart_SubscriptionError(t *testing.T) {
	sub := failingSubscriber{called: make(chan struct{})}
	Start(context.Background(), sub, "test", &fakeSaga{})
	select {
	case <-sub.called:
	case <-time.After(2 * time.Second):
		t.Fatal("Subscribe was not called")
	}
}
