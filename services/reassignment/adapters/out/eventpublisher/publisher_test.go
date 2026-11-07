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
	"github.com/jurwaCST/Jurwa/services/reassignment/ports"
)

var at = time.Date(2026, 11, 7, 10, 0, 0, 0, time.UTC)

func newPublisher(bus eventbus.Publisher) *Publisher {
	n := 0
	return New(bus, func() string { n++; return fmt.Sprintf("id-%d", n) }, func() time.Time { return at })
}

func readOne(t *testing.T, bus *eventbus.Memory) *eventsv1.ReportReassigned {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	got := make(chan *eventsv1.ReportReassigned, 1)
	go func() {
		_ = bus.Subscribe(ctx, eventbus.StreamReportReassigned, "t", "c", func(_ context.Context, b []byte) error {
			var e eventsv1.ReportReassigned
			if err := proto.Unmarshal(b, &e); err != nil {
				return err
			}
			got <- &e
			cancel()
			return nil
		})
	}()
	select {
	case e := <-got:
		return e
	case <-ctx.Done():
		t.Fatal("no report.reassigned")
		return nil
	}
}

func TestPublish_SystemEscalation(t *testing.T) {
	bus := eventbus.NewMemory()
	require.NoError(t, newPublisher(bus).PublishReportReassigned(context.Background(), ports.ReportReassigned{
		ReportID: "r1", FromUserID: "staff", ToUserID: "head", Reason: "SLA lapsed", CorrelationID: "saga-1",
	}))
	e := readOne(t, bus)
	assert.Equal(t, "r1", e.GetReportId())
	assert.Equal(t, "staff", e.GetFromUserId())
	assert.Equal(t, "head", e.GetToUserId())
	assert.Equal(t, "SLA lapsed", e.GetReason())
	assert.Nil(t, e.ActorId, "the system acted")
	assert.Equal(t, "id-1", e.GetMeta().GetEventId())
	assert.Equal(t, "saga-1", e.GetMeta().GetCorrelationId())
	assert.Equal(t, Producer, e.GetMeta().GetProducer())
	assert.Equal(t, at, e.GetMeta().GetOccurredAt().AsTime())
}

func TestPublish_PersonReassigning(t *testing.T) {
	bus := eventbus.NewMemory()
	require.NoError(t, newPublisher(bus).PublishReportReassigned(context.Background(), ports.ReportReassigned{
		ReportID: "r1", FromUserID: "head", ToUserID: "staff", ActorID: "head",
	}))
	e := readOne(t, bus)
	assert.Equal(t, "head", e.GetActorId())
	assert.Equal(t, "id-1", e.GetMeta().GetCorrelationId(), "outside a Saga: its own correlation")
}

type failingBus struct{}

func (failingBus) Publish(context.Context, string, proto.Message) error {
	return errors.New("redis down")
}

func TestPublish_Error(t *testing.T) {
	err := newPublisher(failingBus{}).PublishReportReassigned(context.Background(), ports.ReportReassigned{})
	assert.ErrorContains(t, err, "redis down")
}
