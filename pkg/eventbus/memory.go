package eventbus

import (
	"context"
	"errors"
	"sync"

	"google.golang.org/protobuf/proto"
)

// Memory is an in-memory bus with the same behaviour as the Redis bus: every
// consumer group gets every message, consumers of one group share them, a failing
// handler is tried MaxAttempts times and then the message goes to <stream>.dead.
type Memory struct {
	mu      sync.Mutex
	streams map[string]*memStream
}

type memStream struct {
	entries [][]byte
	groups  map[string]int // group name to index of the next entry to hand out
	changed chan struct{}  // closed and replaced on every new entry
}

// NewMemory returns an empty in-memory bus.
func NewMemory() *Memory {
	return &Memory{streams: map[string]*memStream{}}
}

// Publish adds msg to the stream.
func (m *Memory) Publish(_ context.Context, stream string, msg proto.Message) error {
	if msg == nil {
		return errors.New("eventbus: nil message")
	}
	payload, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	m.add(stream, payload)
	return nil
}

// Subscribe hands every message of the stream to h, sharing them with the other
// consumers of the same group. A new group starts at the first message.
func (m *Memory) Subscribe(ctx context.Context, stream, group, _ string, h Handler) error {
	for {
		payload, ok, wait := m.next(stream, group)
		if ok {
			m.deliver(ctx, stream, payload, h)
			continue
		}
		select {
		case <-ctx.Done():
			return nil
		case <-wait:
		}
	}
}

func (m *Memory) deliver(ctx context.Context, stream string, payload []byte, h Handler) {
	for range MaxAttempts {
		if h(ctx, payload) == nil {
			return
		}
	}
	m.add(stream+DeadSuffix, payload)
}

func (m *Memory) add(stream string, payload []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s := m.stream(stream)
	s.entries = append(s.entries, payload)
	close(s.changed)
	s.changed = make(chan struct{})
}

// next takes the next entry for the group, or returns a channel that is closed
// when a new entry arrives.
func (m *Memory) next(stream, group string) ([]byte, bool, <-chan struct{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s := m.stream(stream)
	i, known := s.groups[group]
	if !known {
		s.groups[group] = 0
	}
	if i < len(s.entries) {
		s.groups[group] = i + 1
		return s.entries[i], true, nil
	}
	return nil, false, s.changed
}

func (m *Memory) stream(name string) *memStream {
	s, ok := m.streams[name]
	if !ok {
		s = &memStream{groups: map[string]int{}, changed: make(chan struct{})}
		m.streams[name] = s
	}
	return s
}
