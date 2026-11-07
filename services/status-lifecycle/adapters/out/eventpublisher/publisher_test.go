package eventpublisher

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/jurwaCST/Jurwa/pkg/eventbus"
	eventsv1 "github.com/jurwaCST/Jurwa/proto/gen/events/v1"
	"github.com/jurwaCST/Jurwa/services/status-lifecycle/domain"
	"github.com/jurwaCST/Jurwa/services/status-lifecycle/ports"
)

var at = time.Date(2026, 11, 7, 10, 0, 0, 0, time.UTC)

func newPublisher(bus eventbus.Publisher) *Publisher {
	n := 0
	return New(bus, func() string { n++; return fmt.Sprintf("id-%d", n) }, func() time.Time { return at })
}

// readOne returns the first message of a stream.
func readOne(t *testing.T, bus *eventbus.Memory, stream string, msg proto.Message) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	done := make(chan error, 1)
	go func() {
		_ = bus.Subscribe(ctx, stream, "t", "c", func(_ context.Context, b []byte) error {
			done <- proto.Unmarshal(b, msg)
			cancel()
			return nil
		})
	}()
	select {
	case err := <-done:
		require.NoError(t, err)
	case <-ctx.Done():
		t.Fatalf("no message on %s", stream)
	}
}

func TestPublishStatusChanged(t *testing.T) {
	bus := eventbus.NewMemory()
	p := newPublisher(bus)
	require.NoError(t, p.PublishStatusChanged(context.Background(), ports.StatusChanged{
		ReportID: "r1", From: domain.Assigned, To: domain.Escalated, Version: 3, CorrelationID: "saga-1",
	}))

	var e eventsv1.StatusChanged
	readOne(t, bus, eventbus.StreamStatusChanged, &e)
	assert.Equal(t, "r1", e.GetReportId())
	assert.Equal(t, "ASSIGNED", e.GetFromStatus().String())
	assert.Equal(t, "ESCALATED", e.GetToStatus().String())
	assert.Equal(t, int32(3), e.GetVersion())
	assert.Nil(t, e.ActorId, "the system acted")
	assert.Equal(t, "id-1", e.GetMeta().GetEventId())
	assert.Equal(t, "saga-1", e.GetMeta().GetCorrelationId())
	assert.Equal(t, Producer, e.GetMeta().GetProducer())
	assert.Equal(t, at, e.GetMeta().GetOccurredAt().AsTime())
}

func TestPublishStatusChanged_UserChangeStartsItsOwnCorrelation(t *testing.T) {
	bus := eventbus.NewMemory()
	p := newPublisher(bus)
	require.NoError(t, p.PublishStatusChanged(context.Background(), ports.StatusChanged{
		ReportID: "r1", From: domain.Open, To: domain.Assigned, ActorID: "u1",
	}))
	var e eventsv1.StatusChanged
	readOne(t, bus, eventbus.StreamStatusChanged, &e)
	assert.Equal(t, "u1", e.GetActorId())
	assert.Equal(t, "id-1", e.GetMeta().GetCorrelationId())
	assert.Equal(t, "id-2", e.GetMeta().GetEventId())
}

func TestPublishEscalationCompensated(t *testing.T) {
	bus := eventbus.NewMemory()
	p := newPublisher(bus)
	require.NoError(t, p.PublishEscalationCompensated(context.Background(), ports.EscalationCompensated{
		ReportID: "r1", FailedStep: "status-lifecycle.set-escalated", UndoneStep: "reassignment.escalate",
		Reason: "RESOLVED to ESCALATED", CorrelationID: "saga-1",
	}))
	var e eventsv1.EscalationCompensated
	readOne(t, bus, eventbus.StreamEscalationCompensated, &e)
	assert.Equal(t, "r1", e.GetReportId())
	assert.Equal(t, "status-lifecycle.set-escalated", e.GetFailedStep())
	assert.Equal(t, "reassignment.escalate", e.GetUndoneStep())
	assert.Equal(t, "RESOLVED to ESCALATED", e.GetReason())
	assert.Equal(t, "saga-1", e.GetMeta().GetCorrelationId())
}

type failingBus struct{}

func (failingBus) Publish(context.Context, string, proto.Message) error {
	return errors.New("redis down")
}

func TestPublish_Errors(t *testing.T) {
	p := newPublisher(failingBus{})
	assert.ErrorContains(t, p.PublishStatusChanged(context.Background(), ports.StatusChanged{}), "redis down")
	assert.ErrorContains(t, p.PublishEscalationCompensated(context.Background(), ports.EscalationCompensated{}), "redis down")
}
