// Package resilience protects gRPC calls between Jurwa services with a timeout,
// a retry with growing waits and a circuit breaker. UnaryClientInterceptor
// combines the three so any client can use them with grpc.WithChainUnaryInterceptor.
// See README.md.
package resilience
