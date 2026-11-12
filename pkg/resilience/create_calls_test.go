package resilience

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	feedbackv1 "github.com/jurwaCST/Jurwa/proto/gen/feedback/v1"
	identityv1 "github.com/jurwaCST/Jurwa/proto/gen/identity/v1"
	internalnotesv1 "github.com/jurwaCST/Jurwa/proto/gen/internal-notes/v1"
	photomediav1 "github.com/jurwaCST/Jurwa/proto/gen/photo-media/v1"
	reassignmentv1 "github.com/jurwaCST/Jurwa/proto/gen/reassignment/v1"
	reportsubmissionv1 "github.com/jurwaCST/Jurwa/proto/gen/report-submission/v1"
	statuslifecyclev1 "github.com/jurwaCST/Jurwa/proto/gen/status-lifecycle/v1"
)

// SCRUM-132: a call that creates or changes something may have worked even
// though it timed out. Retrying it then makes a duplicate (a second report or
// note) or a false error (409 for Register, 409 for a ChangeStatus version).

const register = "/jurwa.identity.v1.IdentityService/Register"

func TestCreateCallIsNotRetriedAfterATimeout(t *testing.T) {
	cfg, _ := testConfig()
	cfg.NoRetryAfterTimeout = map[string]bool{register: true}
	inv := &fakeInvoker{errs: []error{errDeadline, errDeadline}}

	err := UnaryClientInterceptor(cfg)(context.Background(), register, nil, nil, nil, inv.invoke)
	assert.Equal(t, codes.DeadlineExceeded, status.Code(err))
	assert.Equal(t, 1, inv.calls, "one attempt: the first may have created the account")
}

func TestCreateCallIsStillRetriedWhenTheServiceWasUnreachable(t *testing.T) {
	cfg, _ := testConfig()
	cfg.NoRetryAfterTimeout = map[string]bool{register: true}
	inv := &fakeInvoker{errs: []error{errUnavailable, errUnavailable}}

	assert.NoError(t, UnaryClientInterceptor(cfg)(context.Background(), register, nil, nil, nil, inv.invoke))
	assert.Equal(t, 3, inv.calls, "Unavailable: the request never reached the service")
}

func TestReadCallIsStillRetriedAfterATimeout(t *testing.T) {
	cfg, _ := testConfig()
	cfg.NoRetryAfterTimeout = map[string]bool{register: true}
	inv := &fakeInvoker{errs: []error{errDeadline, errDeadline}}

	assert.NoError(t, UnaryClientInterceptor(cfg)(context.Background(), "/jurwa.identity.v1.IdentityService/ValidateToken", nil, nil, nil, inv.invoke))
	assert.Equal(t, 3, inv.calls, "reading twice is harmless")
}

// The defaults name the real generated methods, so a renamed RPC fails here
// instead of silently being retried again.
func TestDefaultConfigListsTheCreateCalls(t *testing.T) {
	want := []string{
		identityv1.IdentityService_Register_FullMethodName,
		reportsubmissionv1.ReportSubmissionService_SubmitReport_FullMethodName,
		photomediav1.PhotoMediaService_UploadPhoto_FullMethodName,
		internalnotesv1.InternalNotesService_AddNote_FullMethodName,
		statuslifecyclev1.StatusLifecycleService_ChangeStatus_FullMethodName,
		reassignmentv1.ReassignmentService_ReassignReport_FullMethodName,
		feedbackv1.FeedbackService_SubmitFeedback_FullMethodName,
	}
	cfg := DefaultConfig()
	for _, m := range want {
		assert.True(t, cfg.NoRetryAfterTimeout[m], m)
	}
	assert.Len(t, cfg.NoRetryAfterTimeout, len(want))
}

func TestRetryableAfterTimeout(t *testing.T) {
	assert.True(t, RetryableUnavailableOnly(errUnavailable))
	assert.False(t, RetryableUnavailableOnly(errDeadline))
	assert.False(t, RetryableUnavailableOnly(ErrCircuitOpen))
	assert.False(t, RetryableUnavailableOnly(errNotFound))
}

func TestRetryUsesItsOwnRetryableWhenSet(t *testing.T) {
	calls := 0
	r := Retry{MaxAttempts: 3, Retryable: RetryableUnavailableOnly, Sleep: func(context.Context, time.Duration) error { return nil }}
	err := r.Do(context.Background(), func(context.Context) error { calls++; return errDeadline })
	assert.Equal(t, codes.DeadlineExceeded, status.Code(err))
	assert.Equal(t, 1, calls)
}
