// Package saga tests the escalation Saga (SCRUM-89) across sla-tracking,
// reassignment and status-lifecycle. The application layers of the three
// services are wired to one in-memory event bus: no gRPC, no Redis.
// Design: docs/design/escalation-saga.md.
package saga

import (
	"context"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/jurwaCST/Jurwa/pkg/eventbus"
	eventsv1 "github.com/jurwaCST/Jurwa/proto/gen/events/v1"
	reevents "github.com/jurwaCST/Jurwa/services/reassignment/adapters/in/events"
	repub "github.com/jurwaCST/Jurwa/services/reassignment/adapters/out/eventpublisher"
	rememory "github.com/jurwaCST/Jurwa/services/reassignment/adapters/out/memory"
	reapp "github.com/jurwaCST/Jurwa/services/reassignment/application"
	redomain "github.com/jurwaCST/Jurwa/services/reassignment/domain"
	stpub "github.com/jurwaCST/Jurwa/services/sla-tracking/adapters/out/eventpublisher"
	stmemory "github.com/jurwaCST/Jurwa/services/sla-tracking/adapters/out/memory"
	stapp "github.com/jurwaCST/Jurwa/services/sla-tracking/application"
	stdomain "github.com/jurwaCST/Jurwa/services/sla-tracking/domain"
	slevents "github.com/jurwaCST/Jurwa/services/status-lifecycle/adapters/in/events"
	slpub "github.com/jurwaCST/Jurwa/services/status-lifecycle/adapters/out/eventpublisher"
	slmemory "github.com/jurwaCST/Jurwa/services/status-lifecycle/adapters/out/memory"
	slapp "github.com/jurwaCST/Jurwa/services/status-lifecycle/application"
	sldomain "github.com/jurwaCST/Jurwa/services/status-lifecycle/domain"
)

const (
	headID  = "22222222-2222-2222-2222-222222222222" // Category Head (DSA), first assignee
	staffID = "33333333-3333-3333-3333-333333333333" // staff member the head handed the report to
)

var t0 = time.Date(2026, 11, 7, 9, 0, 0, 0, time.UTC)

// onePolicy gives every category a 1 hour SLA window.
type onePolicy struct{}

func (onePolicy) Window(context.Context, stdomain.Category) (time.Duration, error) { return time.Hour, nil }

// seen is one event read back from the bus.
type seen struct {
	stream      string
	eventID     string
	correlation string
	at          time.Time
	msg         proto.Message
}

// world is the three services on one bus, plus a recorder of every Saga event.
type world struct {
	bus      *eventbus.Memory
	tracking *stapp.Service
	reassign *reapp.Service
	status   *slapp.Service
	clock    time.Time

	mu     sync.Mutex
	events []seen
}

func newWorld(t *testing.T) *world {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	w := &world{bus: eventbus.NewMemory(), clock: t0}
	now := func() time.Time { return w.clock }

	w.tracking = stapp.NewService(stmemory.New(), onePolicy{}, stpub.New(w.bus, uuid.NewString, time.Now), now)
	w.reassign = reapp.NewService(rememory.New(), repub.New(w.bus, uuid.NewString, time.Now), now, uuid.NewString)
	w.status = slapp.NewService(slmemory.New(), slpub.New(w.bus, uuid.NewString, time.Now), now)

	reevents.Start(ctx, w.bus, "test", w.reassign)
	slevents.Start(ctx, w.bus, "test", w.status)
	w.record(ctx)
	return w
}

// record reads the four Saga streams in its own consumer group.
func (w *world) record(ctx context.Context) {
	streams := map[string]func() proto.Message{
		eventbus.StreamSlaLapsed:             func() proto.Message { return &eventsv1.SlaLapsed{} },
		eventbus.StreamReportReassigned:      func() proto.Message { return &eventsv1.ReportReassigned{} },
		eventbus.StreamStatusChanged:         func() proto.Message { return &eventsv1.StatusChanged{} },
		eventbus.StreamEscalationCompensated: func() proto.Message { return &eventsv1.EscalationCompensated{} },
	}
	for stream, newMsg := range streams {
		go func() {
			_ = w.bus.Subscribe(ctx, stream, "saga-test-recorder", "r", func(_ context.Context, p []byte) error {
				m := newMsg()
				if err := proto.Unmarshal(p, m); err != nil {
					return err
				}
				meta := m.ProtoReflect().Get(m.ProtoReflect().Descriptor().Fields().ByName("meta")).Message().Interface().(*eventsv1.EventMeta)
				w.mu.Lock()
				w.events = append(w.events, seen{stream, meta.GetEventId(), meta.GetCorrelationId(), meta.GetOccurredAt().AsTime(), m})
				w.mu.Unlock()
				return nil
			})
		}()
	}
}

// recorded returns the events seen so far, oldest first.
func (w *world) recorded() []seen {
	w.mu.Lock()
	defer w.mu.Unlock()
	out := append([]seen(nil), w.events...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].at.Before(out[j].at) })
	return out
}

func (w *world) count(stream string) int {
	n := 0
	for _, e := range w.recorded() {
		if e.stream == stream {
			n++
		}
	}
	return n
}

// assignedReport is a report that routing gave to the head, the head handed to
// staff (ASSIGNED), and whose 1 hour SLA timer is running.
func (w *world) assignedReport(t *testing.T) string {
	ctx := context.Background()
	id := uuid.NewString()
	_, err := w.status.Register(ctx, id)
	require.NoError(t, err)
	_, err = w.status.ChangeStatus(ctx, slapp.ChangeStatusCommand{ReportID: id, To: sldomain.Assigned, ActorID: headID})
	require.NoError(t, err)
	_, err = w.reassign.RecordRouted(ctx, id, headID, "DSA")
	require.NoError(t, err)
	_, err = w.reassign.ReassignReport(ctx, reapp.ReassignCommand{
		ReportID: id, ToUserID: staffID, Reason: "Plumbing", ActorID: headID, Role: redomain.RoleCategoryHead,
	})
	require.NoError(t, err)
	_, err = w.tracking.StartTimer(ctx, id, stdomain.Hostel)
	require.NoError(t, err)
	return id
}

// lapse moves the clock past the deadline and runs one escalation check, as the
// sla-escalation-checker function does.
func (w *world) lapse(t *testing.T) {
	w.clock = w.clock.Add(2 * time.Hour)
	_, err := w.tracking.ListDueTimers(context.Background(), time.Time{}, 0)
	require.NoError(t, err)
}

func (w *world) current(t *testing.T, reportID string) redomain.Assignment {
	list, err := w.reassign.ListReassignments(context.Background(), redomain.RoleSystemAdmin, reportID)
	require.NoError(t, err)
	require.NotEmpty(t, list)
	return list[len(list)-1]
}

func (w *world) statusOf(t *testing.T, reportID string) sldomain.ReportStatus {
	rs, err := w.status.GetStatus(context.Background(), reportID)
	require.NoError(t, err)
	return rs
}

const wait, tick = 2 * time.Second, 5 * time.Millisecond

// (a) Happy path: the lapse ends with ESCALATED and the report back with the Category Head.
func TestSaga_LapseEscalatesReportAndHandsItUp(t *testing.T) {
	w := newWorld(t)
	id := w.assignedReport(t)
	w.lapse(t)

	assert.Eventually(t, func() bool {
		return w.statusOf(t, id).Status == sldomain.Escalated && w.current(t, id).Kind == redomain.Escalated
	}, wait, tick)

	rs := w.statusOf(t, id)
	assert.Equal(t, sldomain.Assigned, rs.Previous, "previous status kept for compensation")
	assert.Empty(t, rs.UpdatedBy, "the system acted")
	cur := w.current(t, id)
	assert.Equal(t, headID, cur.AssigneeUserID, "handed up to the Category Head")
	assert.Equal(t, "DSA", cur.AssigneePositionCode)
	assert.Empty(t, cur.AssignedBy, "the system acted")

	assert.Eventually(t, func() bool { return len(w.recorded()) == 3 }, wait, tick)
	ev := w.recorded()
	assert.Equal(t, eventbus.StreamSlaLapsed, ev[0].stream, "the lapse comes first")
	streams := []string{ev[1].stream, ev[2].stream}
	assert.ElementsMatch(t, []string{eventbus.StreamReportReassigned, eventbus.StreamStatusChanged}, streams)
	for _, e := range ev {
		if r, ok := e.msg.(*eventsv1.ReportReassigned); ok {
			assert.Equal(t, staffID, r.GetFromUserId())
			assert.Equal(t, headID, r.GetToUserId())
			assert.Nil(t, r.ActorId, "no actor: the system acted")
		}
		if s, ok := e.msg.(*eventsv1.StatusChanged); ok {
			assert.Equal(t, "ESCALATED", s.GetToStatus().String())
		}
	}
	assert.Zero(t, w.count(eventbus.StreamEscalationCompensated))
}

// (b) A lapse for a report that was RESOLVED meanwhile: the status step fails,
// the compensating event is published and the escalation is undone.
func TestSaga_ResolvedReport_EscalationIsUndone(t *testing.T) {
	w := newWorld(t)
	ctx := context.Background()
	id := w.assignedReport(t)
	for _, s := range []sldomain.Status{sldomain.InProgress, sldomain.Resolved} {
		_, err := w.status.ChangeStatus(ctx, slapp.ChangeStatusCommand{ReportID: id, To: s, ActorID: staffID})
		require.NoError(t, err)
	}
	w.lapse(t) // the timer was not stopped (lost event), so it still lapses

	assert.Eventually(t, func() bool { return w.current(t, id).Kind == redomain.Compensated }, wait, tick)
	cur := w.current(t, id)
	assert.Equal(t, staffID, cur.AssigneeUserID, "the previous assignee is back")
	assert.Equal(t, sldomain.Resolved, w.statusOf(t, id).Status, "no escalation remains")

	assert.Eventually(t, func() bool { return w.count(eventbus.StreamEscalationCompensated) == 1 }, wait, tick)
	for _, e := range w.recorded() {
		if c, ok := e.msg.(*eventsv1.EscalationCompensated); ok {
			assert.Equal(t, "status-lifecycle.set-escalated", c.GetFailedStep())
			assert.Equal(t, "reassignment.escalate", c.GetUndoneStep())
			assert.Equal(t, id, c.GetReportId())
		}
	}
}

// (c) The status step fails for another reason (status-lifecycle does not know
// the report): the same compensation brings the previous assignee back.
func TestSaga_StatusStepFails_PreviousAssigneeIsBack(t *testing.T) {
	w := newWorld(t)
	ctx := context.Background()
	id := uuid.NewString() // known to reassignment and sla-tracking, not to status-lifecycle
	_, err := w.reassign.RecordRouted(ctx, id, headID, "DSA")
	require.NoError(t, err)
	_, err = w.reassign.ReassignReport(ctx, reapp.ReassignCommand{
		ReportID: id, ToUserID: staffID, Reason: "Plumbing", ActorID: headID, Role: redomain.RoleCategoryHead,
	})
	require.NoError(t, err)
	_, err = w.tracking.StartTimer(ctx, id, stdomain.Hostel)
	require.NoError(t, err)
	w.lapse(t)

	assert.Eventually(t, func() bool { return w.current(t, id).Kind == redomain.Compensated }, wait, tick)
	assert.Equal(t, staffID, w.current(t, id).AssigneeUserID)
	list, _ := w.reassign.ListReassignments(ctx, redomain.RoleSystemAdmin, id)
	kinds := []redomain.Kind{}
	for _, a := range list {
		kinds = append(kinds, a.Kind)
	}
	assert.Equal(t, []redomain.Kind{redomain.Routed, redomain.Reassigned, redomain.Escalated, redomain.Compensated}, kinds)
}

// (d) The same sla.lapsed event delivered twice (same event_id) escalates once.
func TestSaga_RepeatedEvent_EscalatesOnce(t *testing.T) {
	w := newWorld(t)
	ctx := context.Background()
	id := w.assignedReport(t)
	other := w.assignedReport(t)

	lapsed := &eventsv1.SlaLapsed{
		Meta:     &eventsv1.EventMeta{EventId: uuid.NewString(), OccurredAt: timestamppb.Now(), Producer: "sla-tracking", CorrelationId: uuid.NewString()},
		ReportId: id, Category: 1, Deadline: timestamppb.New(t0.Add(time.Hour)),
	}
	require.NoError(t, w.bus.Publish(ctx, eventbus.StreamSlaLapsed, lapsed))
	require.NoError(t, w.bus.Publish(ctx, eventbus.StreamSlaLapsed, lapsed))
	// A third, different lapse: once it is handled, both copies before it are too.
	marker := proto.Clone(lapsed).(*eventsv1.SlaLapsed)
	marker.Meta.EventId, marker.ReportId = uuid.NewString(), other
	require.NoError(t, w.bus.Publish(ctx, eventbus.StreamSlaLapsed, marker))

	assert.Eventually(t, func() bool {
		return w.statusOf(t, other).Status == sldomain.Escalated && w.current(t, other).Kind == redomain.Escalated
	}, wait, tick)

	list, _ := w.reassign.ListReassignments(ctx, redomain.RoleSystemAdmin, id)
	escalations := 0
	for _, a := range list {
		if a.Kind == redomain.Escalated {
			escalations++
		}
	}
	assert.Equal(t, 1, escalations, "one escalated assignment")
	assert.Equal(t, int32(3), w.statusOf(t, id).Version, "OPEN 1, ASSIGNED 2, ESCALATED 3: changed once")
}

// (e) Every event of one run carries the correlation_id of the sla.lapsed event.
func TestSaga_OneCorrelationIDThroughTheRun(t *testing.T) {
	w := newWorld(t)
	ctx := context.Background()
	id := w.assignedReport(t)
	for _, s := range []sldomain.Status{sldomain.InProgress, sldomain.Resolved} {
		_, err := w.status.ChangeStatus(ctx, slapp.ChangeStatusCommand{ReportID: id, To: s, ActorID: staffID})
		require.NoError(t, err)
	}
	w.lapse(t)

	// lapse, escalate (reassigned), compensated, restore (reassigned)
	assert.Eventually(t, func() bool { return len(w.recorded()) == 4 }, wait, tick)
	ev := w.recorded()
	require.Equal(t, eventbus.StreamSlaLapsed, ev[0].stream)
	want := ev[0].correlation
	require.NotEmpty(t, want)
	ids := map[string]bool{}
	for _, e := range ev {
		assert.Equal(t, want, e.correlation, e.stream)
		assert.NotEmpty(t, e.eventID, e.stream)
		ids[e.eventID] = true
	}
	assert.Len(t, ids, 4, "every event has its own event_id")
}
