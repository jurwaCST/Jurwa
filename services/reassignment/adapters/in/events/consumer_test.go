package events

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/jurwaCST/Jurwa/pkg/eventbus"
	eventsv1 "github.com/jurwaCST/Jurwa/proto/gen/events/v1"
	"github.com/jurwaCST/Jurwa/services/reassignment/application"
	"github.com/jurwaCST/Jurwa/services/reassignment/domain"
)

type fakeSaga struct {
	mu          sync.Mutex
	routed      []string // "report|assignee|position"
	routedErr   error
	escalated   []application.SagaEvent
	compensated []application.SagaEvent
}

func (f *fakeSaga) RecordRouted(_ context.Context, reportID, assignee, position string) (domain.Assignment, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.routed = append(f.routed, reportID+"|"+assignee+"|"+position)
	return domain.Assignment{}, f.routedErr
}

func (f *fakeSaga) EscalateOnLapse(_ context.Context, ev application.SagaEvent) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.escalated = append(f.escalated, ev)
	return nil
}

func (f *fakeSaga) CompensateEscalation(_ context.Context, ev application.SagaEvent) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.compensated = append(f.compensated, ev)
	return nil
}

func (f *fakeSaga) counts() (int, int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.escalated), len(f.compensated)
}

func TestStart_RoutesBothStreams(t *testing.T) {
	bus := eventbus.NewMemory()
	app := &fakeSaga{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	Start(ctx, bus, "test", app)

	require.NoError(t, bus.Publish(ctx, eventbus.StreamSlaLapsed, &eventsv1.SlaLapsed{
		Meta: &eventsv1.EventMeta{EventId: "e1", CorrelationId: "c1"}, ReportId: "r1",
	}))
	require.NoError(t, bus.Publish(ctx, eventbus.StreamEscalationCompensated, &eventsv1.EscalationCompensated{
		Meta: &eventsv1.EventMeta{EventId: "e2", CorrelationId: "c1"}, ReportId: "r1",
		UndoneStep: "reassignment.escalate", Reason: "RESOLVED",
	}))
	// A compensation that undoes another service's step is not ours.
	require.NoError(t, bus.Publish(ctx, eventbus.StreamEscalationCompensated, &eventsv1.EscalationCompensated{
		Meta: &eventsv1.EventMeta{EventId: "e3"}, ReportId: "r1", UndoneStep: "status-lifecycle.set-escalated",
	}))

	assert.Eventually(t, func() bool { e, c := app.counts(); return e == 1 && c == 1 }, 2*time.Second, 5*time.Millisecond)
	time.Sleep(50 * time.Millisecond) // give the third event time to arrive
	_, c := app.counts()
	assert.Equal(t, 1, c, "only reassignment.escalate is undone here")

	app.mu.Lock()
	defer app.mu.Unlock()
	assert.Equal(t, application.SagaEvent{EventID: "e1", CorrelationID: "c1", ReportID: "r1"}, app.escalated[0])
	assert.Equal(t, application.SagaEvent{EventID: "e2", CorrelationID: "c1", ReportID: "r1", Reason: "RESOLVED"}, app.compensated[0])
}

func TestHandlers_BadPayload(t *testing.T) {
	bad := []byte{0xff, 0xff}
	assert.Error(t, OnSlaLapsed(&fakeSaga{})(context.Background(), bad))
	assert.Error(t, OnEscalationCompensated(&fakeSaga{})(context.Background(), bad))
}

type failingSubscriber struct{ calls chan struct{} }

func (f failingSubscriber) Subscribe(context.Context, string, string, string, eventbus.Handler) error {
	f.calls <- struct{}{}
	return assert.AnError
}

func TestStart_SubscriptionError(t *testing.T) {
	sub := failingSubscriber{calls: make(chan struct{}, 3)}
	Start(context.Background(), sub, "test", &fakeSaga{})
	for range 3 {
		select {
		case <-sub.calls:
		case <-time.After(2 * time.Second):
			t.Fatal("Subscribe was not called for both streams")
		}
	}
}

func TestOnReportRouted_RecordsTheCategoryHead(t *testing.T) {
	bus := eventbus.NewMemory()
	app := &fakeSaga{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	Start(ctx, bus, "test", app)

	require.NoError(t, bus.Publish(ctx, eventbus.StreamReportRouted, &eventsv1.ReportRouted{
		Meta: &eventsv1.EventMeta{EventId: "e1"}, ReportId: "r1", CategoryHeadId: "head", TargetPosition: "DSA",
	}))
	assert.Eventually(t, func() bool {
		app.mu.Lock()
		defer app.mu.Unlock()
		return len(app.routed) == 1
	}, 2*time.Second, 5*time.Millisecond)
	app.mu.Lock()
	defer app.mu.Unlock()
	assert.Equal(t, "r1|head|DSA", app.routed[0])
}

func TestOnReportRouted_RepeatedOrBadEvent(t *testing.T) {
	payload := func() []byte {
		b, _ := proto.Marshal(&eventsv1.ReportRouted{ReportId: "r1", CategoryHeadId: "head"})
		return b
	}()
	assert.NoError(t, OnReportRouted(&fakeSaga{routedErr: domain.ErrAlreadyAssigned})(context.Background(), payload), "already assigned counts as done")
	assert.ErrorIs(t, OnReportRouted(&fakeSaga{routedErr: application.ErrInvalidArgument})(context.Background(), payload), application.ErrInvalidArgument)
	assert.Error(t, OnReportRouted(&fakeSaga{})(context.Background(), []byte{0xff, 0xff}))
}
