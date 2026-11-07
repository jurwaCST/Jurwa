# pkg/resilience

Timeout, retry and circuit breaker for gRPC calls between Jurwa services. Owner: Sonam Zangmo (SCRUM-91).

## The three patterns

| Pattern | File | What it does |
|---|---|---|
| Timeout | `timeout.go` | `WithTimeout` gives a call a deadline. A call slower than the timeout fails with `DeadlineExceeded` instead of hanging the caller. |
| Retry | `retry.go` | `Retry.Do` runs a call again only for temporary errors (`Unavailable`, `DeadlineExceeded`; only `Unavailable` for the calls in `CreateMethods`), at most `MaxAttempts` times, waiting `BaseDelay`, then twice as long each time. Business answers (`InvalidArgument`, `NotFound`, ...) are never retried. It stops early when the caller's context ends. |
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

Used by: routing-engine (routing-rules and department-directory clients), api-gateway (all its clients), feedback.

## Calls that create something (SCRUM-132)

A call can time out on the caller's side after the service has already done the work. Retrying it then repeats the work. For a read that is harmless, but for a call that creates or changes something it is not: a second report, photo or note, or a false 409 (Register finds its own new account, ChangeStatus finds its own new version). The JMeter run in SCRUM-109 showed exactly this: a Register that timed out at the gateway had worked, and its retry got 409.

So the methods in `CreateMethods` are retried **only on `Unavailable`** (the request never reached the service), never after `DeadlineExceeded`. `DefaultConfig()` sets this through `Config.NoRetryAfterTimeout`; everything else keeps the normal retry.

| Method | Repeat after a timeout would |
|---|---|
| identity `Register` | answer 409 email taken |
| report-submission `SubmitReport` | make a second report |
| photo-media `UploadPhoto` | store the photo twice |
| internal-notes `AddNote` | add the note twice |
| status-lifecycle `ChangeStatus` | answer 409 when `expectedVersion` is set |

When a service starts calling another create or change RPC through this package, add its full method name to `CreateMethods`. A test checks the list against the generated `_FullMethodName` constants, so a renamed RPC fails the build instead of quietly being retried.

## Test

```bash
go test ./pkg/resilience/... -cover
```

Tests inject a fake clock into the breaker and a fake sleep into the retry, so the 10 second wait and the retry waits take no real time.
