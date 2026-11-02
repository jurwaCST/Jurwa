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
}

// DefaultConfig is the team default: 2 second timeout per attempt, at most 3
// attempts waiting 100 ms then 200 ms, and a breaker that opens after 5
// failures in a row for 10 seconds. Use one Config per downstream service so
// each service has its own breaker.
func DefaultConfig() Config {
	return Config{
		Timeout: 2 * time.Second,
		Retry:   Retry{MaxAttempts: 3, BaseDelay: 100 * time.Millisecond},
		Breaker: NewBreaker(5, 10*time.Second, time.Now),
	}
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
		return cfg.Retry.Do(ctx, func(ctx context.Context) error {
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
