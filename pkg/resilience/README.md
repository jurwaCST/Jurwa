# pkg/resilience

Timeout, retry and circuit breaker for gRPC calls between Jurwa services. Owner: Sonam Zangmo (SCRUM-91).

## The three patterns

| Pattern | File | What it does |
|---|---|---|
| Timeout | `timeout.go` | `WithTimeout` gives a call a deadline. A call slower than the timeout fails with `DeadlineExceeded` instead of hanging the caller. |
| Retry | `retry.go` | `Retry.Do` runs a call again only for temporary errors (`Unavailable`, `DeadlineExceeded`), at most `MaxAttempts` times, waiting `BaseDelay`, then twice as long each time. Business answers (`InvalidArgument`, `NotFound`, ...) are never retried. It stops early when the caller's context ends. |
| Circuit breaker | `breaker.go` | `Breaker` counts failures in a row (`Unavailable`, `DeadlineExceeded`, `Internal`, `Unknown`, `ResourceExhausted`). After 5 it **opens**: calls are rejected at once with `ErrCircuitOpen` (code `Unavailable`) without calling the service. After 10 seconds it is **half-open** and lets one trial call through: a success **closes** it, a failure **opens** it again for another 10 seconds. |

`UnaryClientInterceptor` (in `interceptor.go`) combines them. For each call, the retry runs attempts. Each attempt asks the breaker first, calls the service with the timeout, then reports the result to the breaker. A call rejected by an open breaker is not retried.

## Use it

```go
import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/jurwaCST/Jurwa/pkg/resilience"
)

conn, err := grpc.NewClient(addr,
	grpc.WithTransportCredentials(insecure.NewCredentials()),
	grpc.WithChainUnaryInterceptor(resilience.UnaryClientInterceptor(resilience.DefaultConfig())),
)
```

`DefaultConfig()` is a 2 second timeout per attempt, 3 attempts (waits of 100 ms and 200 ms) and a breaker of 5 failures / 10 seconds. Call `DefaultConfig()` once per downstream service, so each service gets its own breaker. To change a value, edit the returned `Config` before passing it in.

Used by: routing-engine (routing-rules and department-directory clients), api-gateway (all its clients).

## Test

```bash
go test ./pkg/resilience/... -cover
```

Tests inject a fake clock into the breaker and a fake sleep into the retry, so the 10 second wait and the retry waits take no real time.
