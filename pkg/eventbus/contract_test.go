package eventbus_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/jurwaCST/Jurwa/pkg/eventbus"
	commonv1 "github.com/jurwaCST/Jurwa/proto/gen/common/v1"
	eventsv1 "github.com/jurwaCST/Jurwa/proto/gen/events/v1"
)

// Bus is what every implementation gives: it can publish and subscribe.
type Bus interface {
	eventbus.Publisher
	eventbus.Subscriber
}

const waitFor = 10 * time.Second

// runBusContract runs the same behaviour checks against any Bus.
// The in-memory fake and the Redis bus must both pass it.
func runBusContract(t *testing.T, newBus func(t *testing.T) Bus) {
	t.Run("publish then subscribe delivers the same message", func(t *testing.T) {
		bus := newBus(t)
		stream := streamName(t)
		sent := statusChanged("evt-1", "rep-1")
		require.NoError(t, bus.Publish(context.Background(), stream, sent))

		got := make(chan []byte, 1)
		subscribe(t, bus, stream, "g1", "c1", func(_ context.Context, p []byte) error {
			got <- p
			return nil
		})

		select {
		case p := <-got:
			var back eventsv1.StatusChanged
			require.NoError(t, proto.Unmarshal(p, &back))
			assert.True(t, proto.Equal(sent, &back), "payload changed on the bus")
		case <-time.After(waitFor):
			t.Fatal("message was not delivered")
		}
	})

	t.Run("two consumer groups each get every message", func(t *testing.T) {
		bus := newBus(t)
		stream := streamName(t)
		a, b := newCollector(), newCollector()
		subscribe(t, bus, stream, "group-a", "c1", a.handle)
		subscribe(t, bus, stream, "group-b", "c1", b.handle)

		for i := range 5 {
			require.NoError(t, bus.Publish(context.Background(), stream, statusChanged(fmt.Sprintf("evt-%d", i), "rep")))
		}

		require.Eventually(t, func() bool { return a.count() == 5 && b.count() == 5 }, waitFor, 20*time.Millisecond)
	})

	t.Run("one group with two consumers shares the messages", func(t *testing.T) {
		bus := newBus(t)
		stream := streamName(t)
		shared := newCollector()
		subscribe(t, bus, stream, "workers", "c1", shared.handle)
		subscribe(t, bus, stream, "workers", "c2", shared.handle)

		for i := range 10 {
			require.NoError(t, bus.Publish(context.Background(), stream, statusChanged(fmt.Sprintf("evt-%d", i), "rep")))
		}

		require.Eventually(t, func() bool { return shared.count() == 10 }, waitFor, 20*time.Millisecond)
		time.Sleep(200 * time.Millisecond) // give a duplicate time to show up
		assert.Equal(t, 10, shared.count(), "each message goes to only one consumer of the group")
		assert.Len(t, shared.uniqueEventIDs(t), 10)
	})

	t.Run("failing handler is tried 3 times then the message goes to the dead stream", func(t *testing.T) {
		bus := newBus(t)
		stream := streamName(t)
		var mu sync.Mutex
		attempts := 0
		subscribe(t, bus, stream, "g1", "c1", func(context.Context, []byte) error {
			mu.Lock()
			defer mu.Unlock()
			attempts++
			return errors.New("handler broken")
		})
		dead := newCollector()
		subscribe(t, bus, stream+eventbus.DeadSuffix, "inspect", "c1", dead.handle)

		sent := statusChanged("evt-bad", "rep-9")
		require.NoError(t, bus.Publish(context.Background(), stream, sent))

		require.Eventually(t, func() bool { return dead.count() == 1 }, waitFor, 20*time.Millisecond)
		mu.Lock()
		assert.Equal(t, eventbus.MaxAttempts, attempts)
		mu.Unlock()
		assert.Equal(t, []string{"evt-bad"}, dead.uniqueEventIDs(t))
	})

	t.Run("handler that fails then succeeds is not dead-lettered", func(t *testing.T) {
		bus := newBus(t)
		stream := streamName(t)
		var mu sync.Mutex
		attempts := 0
		done := make(chan struct{})
		subscribe(t, bus, stream, "g1", "c1", func(context.Context, []byte) error {
			mu.Lock()
			defer mu.Unlock()
			attempts++
			if attempts < 2 {
				return errors.New("first try fails")
			}
			close(done)
			return nil
		})
		dead := newCollector()
		subscribe(t, bus, stream+eventbus.DeadSuffix, "inspect", "c1", dead.handle)

		require.NoError(t, bus.Publish(context.Background(), stream, statusChanged("evt-1", "rep-1")))

		select {
		case <-done:
		case <-time.After(waitFor):
			t.Fatal("handler never succeeded")
		}
		time.Sleep(200 * time.Millisecond)
		assert.Equal(t, 0, dead.count())
	})

	t.Run("same event_id twice is passed to the handler twice", func(t *testing.T) {
		bus := newBus(t)
		stream := streamName(t)
		c := newCollector()
		subscribe(t, bus, stream, "g1", "c1", c.handle)

		msg := statusChanged("evt-same", "rep-1")
		require.NoError(t, bus.Publish(context.Background(), stream, msg))
		require.NoError(t, bus.Publish(context.Background(), stream, msg))

		require.Eventually(t, func() bool { return c.count() == 2 }, waitFor, 20*time.Millisecond)
	})

	t.Run("subscribe returns when the context is cancelled", func(t *testing.T) {
		bus := newBus(t)
		ctx, cancel := context.WithCancel(context.Background())
		returned := make(chan error, 1)
		go func() {
			returned <- bus.Subscribe(ctx, streamName(t), "g1", "c1", func(context.Context, []byte) error { return nil })
		}()
		time.Sleep(100 * time.Millisecond)
		cancel()

		select {
		case err := <-returned:
			assert.NoError(t, err)
		case <-time.After(waitFor):
			t.Fatal("subscribe did not stop")
		}
	})

	t.Run("publish of a nil message is an error", func(t *testing.T) {
		bus := newBus(t)
		assert.Error(t, bus.Publish(context.Background(), streamName(t), nil))
	})
}

// subscribe starts a subscriber in the background and stops it when the test ends.
func subscribe(t *testing.T, bus Bus, stream, group, consumer string, h eventbus.Handler) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = bus.Subscribe(ctx, stream, group, consumer, h)
	}()
	t.Cleanup(func() {
		cancel()
		<-done
	})
}

// streamName gives each test its own stream, so tests on one Redis do not mix.
func streamName(t *testing.T) string {
	return "test:" + strings.NewReplacer(" ", "-", "/", ".").Replace(t.Name())
}

func statusChanged(eventID, reportID string) *eventsv1.StatusChanged {
	return &eventsv1.StatusChanged{
		Meta:       &eventsv1.EventMeta{EventId: eventID, Producer: "test"},
		ReportId:   reportID,
		FromStatus: commonv1.Status_OPEN,
		ToStatus:   commonv1.Status_ASSIGNED,
		Version:    1,
	}
}

// collector is a handler that remembers every payload it gets.
type collector struct {
	mu       sync.Mutex
	payloads [][]byte
}

func newCollector() *collector { return &collector{} }

func (c *collector) handle(_ context.Context, p []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.payloads = append(c.payloads, p)
	return nil
}

func (c *collector) count() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.payloads)
}

func (c *collector) uniqueEventIDs(t *testing.T) []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	seen := map[string]bool{}
	var ids []string
	for _, p := range c.payloads {
		var m eventsv1.StatusChanged
		require.NoError(t, proto.Unmarshal(p, &m))
		if id := m.GetMeta().GetEventId(); !seen[id] {
			seen[id] = true
			ids = append(ids, id)
		}
	}
	return ids
}
