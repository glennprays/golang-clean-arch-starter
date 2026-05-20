// Package logctx threads the request's trace ID through context.Context.
//
// The Fiber HTTP layer stores the trace ID in c.Locals, but usecases,
// services, and repositories only see context.Context. This package
// is the bridge: handlers (via middleware) put the ID in, anything
// downstream pulls it out for log correlation.
package logctx

import "context"

type traceIDKey struct{}

// WithTraceID returns ctx augmented with the given trace ID.
func WithTraceID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, traceIDKey{}, id)
}

// TraceID returns the trace ID from ctx, or "" if absent.
func TraceID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	id, _ := ctx.Value(traceIDKey{}).(string)
	return id
}
