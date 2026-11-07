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
	"github.com/jurwaCST/Jurwa/services/sla-tracking/domain"
	"github.com/jurwaCST/Jurwa/services/sla-tracking/ports"
)

func TestPublishSLALapsed(t *testing.T) {
	bus := eventbus.NewMemory()
	n := 0
	at := time.Date(2026, 11, 7, 10, 0, 0, 0, time.UTC)
	p := New(bus, func() string { n++; return fmt.Sprintf("id-%d", n) }, func() time.Time { return at })
	deadline := at.Add(-time.Minute)

	require.NoError(t, p.PublishSLALapsed(context.Background(), ports.SLALapsed{
		ReportID: "r1", Category: domain.AcademicBlock, Deadline: deadline, LapsedAt: at,
	}))

	got := make(chan *eventsv1.SlaLapsed, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		_ = bus.Subscribe(ctx, eventbus.StreamSlaLapsed, "t", "c", func(_ context.Context, b []byte) error {
			var e eventsv1.SlaLapsed
			if err := proto.Unmarshal(b, &e); err != nil {
				return err
			}
			got <- &e
			return nil
		})
	}()
	e := <-got
	assert.Equal(t, "r1", e.GetReportId())
	assert.Equal(t, "ACADEMIC_BLOCK", e.GetCategory().String())
	assert.Equal(t, deadline, e.GetDeadline().AsTime())
	assert.Equal(t, "id-1", e.GetMeta().GetEventId())
	assert.Equal(t, "id-2", e.GetMeta().GetCorrelationId(), "a new Saga run")
	assert.Equal(t, Producer, e.GetMeta().GetProducer())
	assert.Equal(t, at, e.GetMeta().GetOccurredAt().AsTime())
}

type failingBus struct{}

func (failingBus) Publish(context.Context, string, proto.Message) error {
	return errors.New("redis down")
}

func TestPublishSLALapsed_Error(t *testing.T) {
	p := New(failingBus{}, func() string { return "x" }, time.Now)
	assert.ErrorContains(t, p.PublishSLALapsed(context.Background(), ports.SLALapsed{}), "redis down")
}
