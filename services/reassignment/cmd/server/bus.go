package main

import (
	"context"
	"os"
	"time"

	"github.com/jurwaCST/Jurwa/pkg/eventbus"
)

// eventBus connects to Redis Streams when REDIS_URL is set. Without it the
// service keeps its log publisher and does not take part in the escalation Saga,
// so `make run` and the unit tests need no Redis.
func eventBus(ctx context.Context) (*eventbus.Redis, error) {
	url := os.Getenv("REDIS_URL")
	if url == "" {
		return nil, nil
	}
	bus, err := eventbus.NewRedis(url)
	if err != nil {
		return nil, err
	}
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := bus.Ping(pingCtx); err != nil {
		_ = bus.Close()
		return nil, err
	}
	return bus, nil
}

// consumerName names this instance in its consumer group: the pod name in
// Kubernetes, or the service name.
func consumerName(fallback string) string {
	if h, err := os.Hostname(); err == nil && h != "" {
		return h
	}
	return fallback
}
