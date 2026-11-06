package observability

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc/metadata"
)

// CorrelationHeader is the gRPC metadata key that carries one request's id
// across services, so their log lines can be matched up.
const CorrelationHeader = "x-correlation-id"

type correlationKey struct{}

// withCorrelationID takes the id from the incoming metadata, or makes a new
// one, and stores it in the context for the handler.
func withCorrelationID(ctx context.Context) (context.Context, string) {
	id := ""
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if v := md.Get(CorrelationHeader); len(v) > 0 {
			id = v[0]
		}
	}
	if id == "" {
		id = uuid.NewString()
	}
	return context.WithValue(ctx, correlationKey{}, id), id
}

// CorrelationID is the current request's id, "" outside a request.
func CorrelationID(ctx context.Context) string {
	id, _ := ctx.Value(correlationKey{}).(string)
	return id
}
