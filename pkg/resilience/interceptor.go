package resilience

import (
	"context"
	"time"

	"google.golang.org/grpc"
)

// Config chooses the patterns UnaryClientInterceptor applies.
type Config struct {
	Timeout time.Duration // per attempt; 0 means no timeout
	Retry   Retry
	Breaker *Breaker // nil means no circuit breaker
	// NoRetryAfterTimeout lists full gRPC method names that create or change
	// something. They are retried only on Unavailable, never after a timeout
	// (SCRUM-132). DefaultConfig fills it with CreateMethods.
	NoRetryAfterTimeout map[string]bool
}

// CreateMethods are the RPCs called through this package that are not safe to
// repeat after a timeout: a repeat can create a second report, photo or note, or
// answer 409 (Register: email taken; ChangeStatus: version changed) although the
// first attempt worked. Add a method here when a service starts calling one.
var CreateMethods = []string{
	"/jurwa.identity.v1.IdentityService/Register",
	"/jurwa.reportsubmission.v1.ReportSubmissionService/SubmitReport",
	"/jurwa.photomedia.v1.PhotoMediaService/UploadPhoto",
	"/jurwa.internalnotes.v1.InternalNotesService/AddNote",
	"/jurwa.statuslifecycle.v1.StatusLifecycleService/ChangeStatus",
	"/jurwa.reassignment.v1.ReassignmentService/ReassignReport",
	"/jurwa.feedback.v1.FeedbackService/SubmitFeedback",
}

// DefaultConfig is the team default: 2 second timeout per attempt, at most 3
// attempts waiting 100 ms then 200 ms (CreateMethods only on Unavailable), and a
// breaker that opens after 5 failures in a row for 10 seconds. Use one Config per downstream service so
// each service has its own breaker.
func DefaultConfig() Config {
	return Config{
		Timeout:             2 * time.Second,
		Retry:               Retry{MaxAttempts: 3, BaseDelay: 100 * time.Millisecond},
		Breaker:             NewBreaker(5, 10*time.Second, time.Now),
		NoRetryAfterTimeout: noRetryAfterTimeout(),
	}
}

func noRetryAfterTimeout() map[string]bool {
	m := make(map[string]bool, len(CreateMethods))
	for _, method := range CreateMethods {
		m[method] = true
	}
	return m
}

// UnaryClientInterceptor wraps every unary call: the retry runs attempts, each
// attempt first asks the breaker, then calls the service with a timeout, and
// the result is reported back to the breaker.
//
//	conn, err := grpc.NewClient(addr,
//		grpc.WithTransportCredentials(insecure.NewCredentials()),
//		grpc.WithChainUnaryInterceptor(resilience.UnaryClientInterceptor(resilience.DefaultConfig())))
func UnaryClientInterceptor(cfg Config) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		retry := cfg.Retry
		if cfg.NoRetryAfterTimeout[method] {
			retry.Retryable = RetryableUnavailableOnly
		}
		return retry.Do(ctx, func(ctx context.Context) error {
			if cfg.Breaker != nil {
				if err := cfg.Breaker.Allow(); err != nil {
					return err
				}
			}
			err := WithTimeout(ctx, cfg.Timeout, func(ctx context.Context) error {
				return invoker(ctx, method, req, reply, cc, opts...)
			})
			if cfg.Breaker != nil {
				cfg.Breaker.Record(err)
			}
			return err
		})
	}
}
