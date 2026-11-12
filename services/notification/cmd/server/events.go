package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/jurwaCST/Jurwa/pkg/eventbus"
	"github.com/jurwaCST/Jurwa/pkg/resilience"
	reportsubmissionv1 "github.com/jurwaCST/Jurwa/proto/gen/report-submission/v1"
	"github.com/jurwaCST/Jurwa/services/notification/adapters/in/events"
	"github.com/jurwaCST/Jurwa/services/notification/adapters/out/reportclient"
	"github.com/jurwaCST/Jurwa/services/notification/application"
)

// startConsumers reads report.submitted and status.changed when REDIS_URL is
// set (SCRUM-155). With the store in memory every stream is read again from
// its first event on start (Rewind), so messages come back after a restart;
// events seen before are skipped by their id. With PostgreSQL the group
// carries on where it stopped.
func startConsumers(ctx context.Context, app *application.Service, serviceName string, inMemory bool) error {
	url := os.Getenv("REDIS_URL")
	if url == "" {
		slog.Info("events: not read (REDIS_URL not set)")
		return nil
	}
	bus, err := eventbus.NewRedis(url)
	if err != nil {
		return err
	}
	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := bus.Ping(pingCtx); err != nil {
		_ = bus.Close()
		return err
	}
	if inMemory {
		for _, stream := range events.Streams() {
			if err := bus.Rewind(ctx, stream, events.Group); err != nil {
				_ = bus.Close()
				return err
			}
		}
	}

	// status.changed does not name the student, so the report is looked up.
	conn, err := grpc.NewClient(envOr("REPORT_SUBMISSION_ADDR", reportclient.DefaultAddr),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(resilience.UnaryClientInterceptor(resilience.DefaultConfig())),
	)
	if err != nil {
		_ = bus.Close()
		return err
	}
	reports := reportclient.New(reportsubmissionv1.NewReportSubmissionServiceClient(conn))

	consumer, _ := os.Hostname()
	if consumer == "" {
		consumer = serviceName
	}
	go func() {
		defer bus.Close()
		defer conn.Close()
		if err := events.NewConsumer(app, reports).Run(ctx, bus, consumer); err != nil {
			slog.Error("event consumer stopped", "error", err)
			os.Exit(1)
		}
	}()
	slog.Info("events: reading", "streams", events.Streams(), "group", events.Group, "consumer", consumer)
	return nil
}
