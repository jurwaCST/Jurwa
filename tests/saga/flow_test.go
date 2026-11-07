package saga

// The whole flow through events, as in the demo (SCRUM-89, demo step 6):
// report-submission publishes report.submitted and routing-engine publishes
// report.routed; status-lifecycle, reassignment and sla-tracking react to them
// without any seed setting. A resolved report stops its timer through
// status.changed.

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/jurwaCST/Jurwa/pkg/eventbus"
	commonv1 "github.com/jurwaCST/Jurwa/proto/gen/common/v1"
	eventsv1 "github.com/jurwaCST/Jurwa/proto/gen/events/v1"
	redomain "github.com/jurwaCST/Jurwa/services/reassignment/domain"
	stevents "github.com/jurwaCST/Jurwa/services/sla-tracking/adapters/in/events"
	stdomain "github.com/jurwaCST/Jurwa/services/sla-tracking/domain"
	slapp "github.com/jurwaCST/Jurwa/services/status-lifecycle/application"
	sldomain "github.com/jurwaCST/Jurwa/services/status-lifecycle/domain"
)

// newFullWorld also lets sla-tracking read report.routed and status.changed.
func newFullWorld(t *testing.T) *world {
	w := newWorld(t)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	stevents.Start(ctx, w.bus, "test", w.tracking)
	return w
}

func meta(producer string) *eventsv1.EventMeta {
	return &eventsv1.EventMeta{EventId: uuid.NewString(), OccurredAt: timestamppb.Now(), Producer: producer, CorrelationId: uuid.NewString()}
}

func (w *world) submitted(t *testing.T, id string) {
	require.NoError(t, w.bus.Publish(context.Background(), eventbus.StreamReportSubmitted, &eventsv1.ReportSubmitted{
		Meta: meta("report-submission"),
		Report: &commonv1.Report{
			Id: id, ReporterId: uuid.NewString(), Category: commonv1.Category_HOSTEL,
			Location: &commonv1.Location{Building: "Hostel A"}, Description: "Tap leaking", Status: commonv1.Status_OPEN,
		},
	}))
}

func (w *world) routed(t *testing.T, id string) {
	require.NoError(t, w.bus.Publish(context.Background(), eventbus.StreamReportRouted, &eventsv1.ReportRouted{
		Meta: meta("routing-engine"), ReportId: id, Category: commonv1.Category_HOSTEL,
		CategoryHeadId: headID, TargetPosition: "DSA", RuleId: "rule-1",
	}))
}

func (w *world) timer(id string) (stdomain.Timer, bool) {
	tm, err := w.tracking.GetTimer(context.Background(), id)
	return tm, err == nil
}

func (w *world) statusIs(id string, s sldomain.Status) bool {
	rs, err := w.status.GetStatus(context.Background(), id)
	return err == nil && rs.Status == s
}

func (w *world) currentIs(id string, kind redomain.Kind, user string) bool {
	list, err := w.reassign.ListReassignments(context.Background(), redomain.RoleSystemAdmin, id)
	if err != nil || len(list) == 0 {
		return false
	}
	cur := list[len(list)-1]
	return cur.Kind == kind && cur.AssigneeUserID == user
}

// Submitted and routed reports get a status, a first assignee and a running
// SLA timer from events alone; when the SLA lapses the report is escalated.
func TestFlow_SubmittedAndRoutedReportIsEscalatedAfterItsSLA(t *testing.T) {
	w := newFullWorld(t)
	id := uuid.NewString()
	w.submitted(t, id)
	w.routed(t, id)

	assert.Eventually(t, func() bool {
		tm, ok := w.timer(id)
		return w.statusIs(id, sldomain.Assigned) && w.currentIs(id, redomain.Routed, headID) && ok && tm.State == stdomain.Running
	}, wait, tick, "ASSIGNED, routed to the head, timer running")
	tm, _ := w.timer(id)
	assert.Equal(t, t0.Add(time.Hour), tm.Deadline, "deadline from the sla-policy window")

	w.lapse(t)
	assert.Eventually(t, func() bool {
		return w.statusIs(id, sldomain.Escalated) && w.currentIs(id, redomain.Escalated, headID)
	}, wait, tick, "escalated after the lapse")
}

// Resolving a report publishes status.changed, which stops its timer, so the
// report is never escalated.
func TestFlow_ResolvedReportStopsItsTimer(t *testing.T) {
	w := newFullWorld(t)
	ctx := context.Background()
	id := uuid.NewString()
	w.submitted(t, id)
	w.routed(t, id)
	assert.Eventually(t, func() bool { _, ok := w.timer(id); return ok && w.statusIs(id, sldomain.Assigned) }, wait, tick)

	for _, s := range []sldomain.Status{sldomain.InProgress, sldomain.Resolved} {
		_, err := w.status.ChangeStatus(ctx, slapp.ChangeStatusCommand{ReportID: id, To: s, ActorID: headID})
		require.NoError(t, err)
	}
	assert.Eventually(t, func() bool { tm, _ := w.timer(id); return tm.State == stdomain.Stopped }, wait, tick, "timer stopped")

	w.lapse(t)
	time.Sleep(50 * time.Millisecond)
	assert.True(t, w.statusIs(id, sldomain.Resolved), "never escalated")
	assert.Zero(t, w.count(eventbus.StreamSlaLapsed))
}

// report.routed can be read before report.submitted (two streams, read in
// parallel). The report still ends up ASSIGNED with a timer.
func TestFlow_RoutedBeforeSubmitted(t *testing.T) {
	w := newFullWorld(t)
	id := uuid.NewString()
	w.routed(t, id)
	w.submitted(t, id)

	assert.Eventually(t, func() bool {
		_, ok := w.timer(id)
		return w.statusIs(id, sldomain.Assigned) && w.currentIs(id, redomain.Routed, headID) && ok
	}, wait, tick)
	rs, err := w.status.GetStatus(context.Background(), id)
	require.NoError(t, err)
	assert.Equal(t, int32(2), rs.Version, "OPEN then ASSIGNED, once")
}

// The same report.routed twice changes nothing the second time.
func TestFlow_RepeatedRoutedEvent(t *testing.T) {
	w := newFullWorld(t)
	id := uuid.NewString()
	w.submitted(t, id)
	w.routed(t, id)
	w.routed(t, id)
	marker := uuid.NewString()
	w.submitted(t, marker)
	w.routed(t, marker)

	assert.Eventually(t, func() bool { _, ok := w.timer(marker); return ok && w.statusIs(marker, sldomain.Assigned) }, wait, tick)
	time.Sleep(50 * time.Millisecond)
	rs, _ := w.status.GetStatus(context.Background(), id)
	assert.Equal(t, int32(2), rs.Version)
	list, _ := w.reassign.ListReassignments(context.Background(), redomain.RoleSystemAdmin, id)
	assert.Len(t, list, 1, "one ROUTED assignment")
}
