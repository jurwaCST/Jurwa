package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/jurwaCST/Jurwa/pkg/eventbus"
	"github.com/jurwaCST/Jurwa/services/analytics/adapters/in/events"
	"github.com/jurwaCST/Jurwa/services/analytics/application"
)

// waitForReport is how long a status change waits for its report (SCRUM-160).
const waitForReport = 5 * time.Second

// startConsumers reads report.submitted and status.changed when REDIS_URL is
// set (SCRUM-160), so new reports and status changes reach the dashboards.
// With the store in memory every stream is read again from its first event on
// start (Rewind), so the counts come back after a restart; events seen before
// are skipped by their id. With PostgreSQL the group carries on where it stopped.
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

	consumer, _ := os.Hostname()
	if consumer == "" {
		consumer = serviceName
	}
	go func() {
		defer bus.Close()
		if err := events.NewConsumer(app, waitForReport).Run(ctx, bus, consumer); err != nil {
			slog.Error("event consumer stopped", "error", err)
			os.Exit(1)
		}
	}()
	slog.Info("events: reading", "streams", events.Streams(), "group", events.Group, "consumer", consumer)
	return nil
}
