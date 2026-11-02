package main

import (
	"fmt"
	"time"

	"github.com/jurwaCST/Jurwa/services/analytics/domain"
)

// demoEvents returns a fixed set of events for demos, used only when
// SEED_DEMO_DATA=true. In Sprint 3 the Redis Streams consumer feeds real
// events instead. Six reports, created over the last two days:
//
//	hostel-1    HOSTEL          OPEN
//	hostel-2    HOSTEL          RESOLVED after 6 hours
//	academic-1  ACADEMIC_BLOCK  IN_PROGRESS
//	academic-2  ACADEMIC_BLOCK  ESCALATED (SLA lapsed)
//	others-1    OTHERS          ASSIGNED
//	others-2    OTHERS          RESOLVED after 12 hours
func demoEvents(now time.Time) []domain.Event {
	start := now.Add(-48 * time.Hour)
	n := 0
	ev := func(t domain.EventType, report string, c domain.Category, to domain.Status, after time.Duration) domain.Event {
		n++
		return domain.Event{
			ID:         fmt.Sprintf("demo-%02d", n),
			Type:       t,
			ReportID:   report,
			Category:   c,
			ToStatus:   to,
			OccurredAt: start.Add(after),
		}
	}
	sub := func(report string, c domain.Category) domain.Event {
		return ev(domain.ReportSubmitted, report, c, domain.StatusUnspecified, 0)
	}
	chg := func(report string, to domain.Status, after time.Duration) domain.Event {
		return ev(domain.StatusChanged, report, domain.CategoryUnspecified, to, after)
	}
	return []domain.Event{
		sub("hostel-1", domain.Hostel),
		sub("hostel-2", domain.Hostel),
		sub("academic-1", domain.AcademicBlock),
		sub("academic-2", domain.AcademicBlock),
		sub("others-1", domain.Others),
		sub("others-2", domain.Others),
		chg("hostel-2", domain.Assigned, time.Hour),
		chg("hostel-2", domain.InProgress, 2*time.Hour),
		chg("hostel-2", domain.Resolved, 6*time.Hour),
		chg("academic-1", domain.Assigned, time.Hour),
		chg("academic-1", domain.InProgress, 3*time.Hour),
		chg("academic-2", domain.Escalated, 47*time.Hour),
		chg("others-1", domain.Assigned, time.Hour),
		chg("others-2", domain.Assigned, time.Hour),
		chg("others-2", domain.Resolved, 12*time.Hour),
	}
}
