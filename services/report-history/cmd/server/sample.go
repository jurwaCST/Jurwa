package main

import (
	"context"
	"time"

	"github.com/jurwaCST/Jurwa/services/report-history/application"
	"github.com/jurwaCST/Jurwa/services/report-history/domain"
)

// Sample ids, so grpcurl calls in the README always work after SEED_SAMPLE=true.
const (
	sampleStudentID = "1b2c3d4e-5f60-4a7b-8c9d-0e1f2a3b4c5d"
	sampleDSAID     = "4e5f6071-8293-4d0e-9f2a-3b4c5d6e7f80"
	sampleReport1   = "6f1c2a7e-3b4d-4c5e-8f90-1a2b3c4d5e6f"
	sampleReport2   = "9a8b7c6d-5e4f-4a3b-8c2d-1e0f9a8b7c6d"
)

// seedSample records two reports of one student, for local demos only.
// In Sprint 3 the data comes from Redis Streams events instead.
func seedSample(ctx context.Context, app *application.Service) error {
	t0 := time.Date(2026, 10, 30, 9, 0, 0, 0, time.UTC)
	reports := []application.SubmittedInput{
		{ReportID: sampleReport1, ReporterID: sampleStudentID, Category: domain.CategoryHostel,
			Location: domain.Location{Building: "Hostel A", Wing: "Wing 1", Area: "Bathroom 1"}, CreatedAt: t0},
		{ReportID: sampleReport2, ReporterID: sampleStudentID, Category: domain.CategoryOthers,
			Location: domain.Location{Building: "Campus", Area: "Wi-Fi"}, CreatedAt: t0.Add(24 * time.Hour)},
	}
	for _, r := range reports {
		if _, err := app.RecordSubmitted(ctx, r); err != nil {
			return err
		}
	}
	changes := []domain.EntryInput{
		{ReportID: sampleReport1, EventType: domain.EventStatusChanged, FromStatus: domain.StatusOpen, ToStatus: domain.StatusAssigned,
			Summary: "Assigned to DSA", OccurredAt: t0.Add(time.Minute)},
		{ReportID: sampleReport1, EventType: domain.EventStatusChanged, FromStatus: domain.StatusAssigned, ToStatus: domain.StatusInProgress,
			ActorID: sampleDSAID, Summary: "Plumber is on the way", OccurredAt: t0.Add(3 * time.Hour)},
	}
	for _, c := range changes {
		if _, err := app.Record(ctx, c); err != nil {
			return err
		}
	}
	return nil
}
