package eventbus

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/proto"
)

const payloadField = "payload"

// Redis is the event bus on Redis Streams.
type Redis struct {
	client *redis.Client
	// Block is how long one XREADGROUP waits for new messages.
	Block time.Duration
	// RetryDelay is the pause between two attempts of a failing handler.
	RetryDelay time.Duration
}

// NewRedis connects to Redis, for example "redis://:PASSWORD@localhost:6379/0".
func NewRedis(url string) (*Redis, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("eventbus: bad REDIS_URL: %w", err)
	}
	return &Redis{client: redis.NewClient(opts), Block: 2 * time.Second, RetryDelay: 100 * time.Millisecond}, nil
}

// Ping checks that Redis answers.
func (r *Redis) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// Close closes the connection.
func (r *Redis) Close() error {
	return r.client.Close()
}

// Publish adds msg to the stream with XADD <stream> * payload <bytes>.
func (r *Redis) Publish(ctx context.Context, stream string, msg proto.Message) error {
	if msg == nil {
		return errors.New("eventbus: nil message")
	}
	payload, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	return r.client.XAdd(ctx, &redis.XAddArgs{Stream: stream, Values: map[string]any{payloadField: payload}}).Err()
}

// Subscribe creates the group if needed (starting at the first entry, so events
// published before a service starts are not lost), first finishes this consumer's
// unacknowledged messages, then reads new ones until ctx is cancelled.
func (r *Redis) Subscribe(ctx context.Context, stream, group, consumer string, h Handler) error {
	err := r.client.XGroupCreateMkStream(ctx, stream, group, "0").Err()
	if err != nil && !strings.HasPrefix(err.Error(), "BUSYGROUP") {
		return stopped(ctx, err)
	}

	// "0" returns messages delivered to this consumer before but never acked
	// (for example after a crash); ">" returns new messages.
	start := "0"
	for {
		res, err := r.client.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group: group, Consumer: consumer, Streams: []string{stream, start}, Count: 10, Block: r.Block,
		}).Result()
		if errors.Is(err, redis.Nil) {
			continue
		}
		if err != nil {
			return stopped(ctx, err)
		}
		msgs := res[0].Messages
		if start == "0" && len(msgs) == 0 {
			start = ">"
			continue
		}
		for _, m := range msgs {
			if err := r.handle(ctx, stream, group, m, h); err != nil {
				return stopped(ctx, err)
			}
		}
	}
}

func (r *Redis) handle(ctx context.Context, stream, group string, m redis.XMessage, h Handler) error {
	raw, ok := m.Values[payloadField].(string)
	lastErr := errors.New("entry has no payload field")
	if ok {
		for attempt := 1; attempt <= MaxAttempts; attempt++ {
			if lastErr = h(ctx, []byte(raw)); lastErr == nil {
				return r.client.XAck(ctx, stream, group, m.ID).Err()
			}
			if attempt < MaxAttempts {
				select {
				case <-ctx.Done():
					return ctx.Err() // not acked, so it is read again on restart
				case <-time.After(r.RetryDelay):
				}
			}
		}
	}

	err := r.client.XAdd(ctx, &redis.XAddArgs{Stream: stream + DeadSuffix, Values: map[string]any{
		payloadField: raw, "error": lastErr.Error(), "group": group, "source_id": m.ID,
	}}).Err()
	if err != nil {
		return err
	}
	return r.client.XAck(ctx, stream, group, m.ID).Err()
}

// stopped turns an error caused by cancelling ctx into a clean stop.
func stopped(ctx context.Context, err error) error {
	if ctx.Err() != nil {
		return nil
	}
	return err
}
