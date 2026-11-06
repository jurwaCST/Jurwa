// Package eventbus is the shared event bus of Jurwa. Services publish protobuf
// events to Redis Streams and read them in consumer groups named after themselves.
// Unit tests use NewMemory, which behaves the same without Redis.
package eventbus

import (
	"context"

	"google.golang.org/protobuf/proto"
)

// The six streams, see docs/design/database-schema.md section 6.
const (
	StreamReportSubmitted       = "events:report.submitted"
	StreamReportRouted          = "events:report.routed"
	StreamStatusChanged         = "events:status.changed"
	StreamReportReassigned      = "events:report.reassigned"
	StreamSlaLapsed             = "events:sla.lapsed"
	StreamEscalationCompensated = "events:escalation.compensated"
)

const (
	// MaxAttempts is how many times a handler is called for one message before
	// the message is moved to the dead stream.
	MaxAttempts = 3
	// DeadSuffix is added to a stream name to get its dead-letter stream.
	DeadSuffix = ".dead"
)

// Publisher sends an event to a stream.
type Publisher interface {
	Publish(ctx context.Context, stream string, msg proto.Message) error
}

// Handler processes one event payload (protobuf bytes). It must be safe to run
// twice for the same event_id: the bus can deliver a message more than once.
type Handler func(ctx context.Context, payload []byte) error

// Subscriber reads a stream in a consumer group. Subscribe blocks until ctx is
// cancelled, then returns nil.
type Subscriber interface {
	Subscribe(ctx context.Context, stream, group, consumer string, h Handler) error
}
