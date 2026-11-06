package resilience_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/jurwaCST/Jurwa/pkg/resilience"
	"github.com/jurwaCST/Jurwa/pkg/resilience/mocks"
)

// Interaction tests: gomock checks how many times Retry runs the call. The
// circuit breaker keeps its hand-written fake clock (resilience_test.go).

func newRetry(waits *[]time.Duration) resilience.Retry {
	return resilience.Retry{
		MaxAttempts: 3,
		BaseDelay:   100 * time.Millisecond,
		Sleep: func(_ context.Context, d time.Duration) error {
			*waits = append(*waits, d)
			return nil
		},
	}
}

func TestRetryCallsAFailingCallExactlyThreeTimes(t *testing.T) {
	ctrl := gomock.NewController(t)
	caller := mocks.NewMockCaller(ctrl)
	unavailable := status.Error(codes.Unavailable, "down")
	caller.EXPECT().Call(gomock.Any()).Return(unavailable).Times(3)

	var waits []time.Duration
	err := newRetry(&waits).Do(context.Background(), caller.Call)

	assert.Equal(t, codes.Unavailable, status.Code(err))
	assert.Equal(t, []time.Duration{100 * time.Millisecond, 200 * time.Millisecond}, waits)
}

func TestRetryCallsANonRetryableErrorOnlyOnce(t *testing.T) {
	ctrl := gomock.NewController(t)
	caller := mocks.NewMockCaller(ctrl)
	caller.EXPECT().Call(gomock.Any()).Return(status.Error(codes.InvalidArgument, "bad input")).Times(1)

	var waits []time.Duration
	err := newRetry(&waits).Do(context.Background(), caller.Call)

	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Empty(t, waits, "no wait when nothing is retried")
}

func TestRetryStopsAfterTheFirstSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	caller := mocks.NewMockCaller(ctrl)
	unavailable := status.Error(codes.Unavailable, "down")
	gomock.InOrder(
		caller.EXPECT().Call(gomock.Any()).Return(unavailable).Times(2),
		caller.EXPECT().Call(gomock.Any()).Return(nil).Times(1),
	)

	var waits []time.Duration
	err := newRetry(&waits).Do(context.Background(), caller.Call)

	assert.NoError(t, err)
	assert.Len(t, waits, 2)
}
