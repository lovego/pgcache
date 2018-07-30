package tracer

import (
	"context"
)

var spanKey spanKeyStruct

type spanKeyStruct struct {
}

// Context return a tracer context with the span
func Context(ctx context.Context, s *Span) context.Context {
	if s == nil {
		return ctx
	}
	return context.WithValue(ctx, spanKey, s)
}

// StartContext start a new tracer context on the given tracer context
func StartContext(ctx context.Context, name string) context.Context {
	return context.WithValue(context.Background(), spanKey, StartSpan(ctx, name))
}

// GetSpan get span from a tracer context
func GetSpan(ctx context.Context) *Span {
	if ctx != nil {
		if v := ctx.Value(spanKey); v != nil {
			if s, ok := v.(*Span); ok {
				return s
			}
		}
	}
	return nil
}

// Tag add a tag to a tracer context
func Tag(ctx context.Context, k string, v interface{}) {
	GetSpan(ctx).Tag(k, v)
}

// DebugTag add a tag to a tracer context if debug is enabled
func DebugTag(ctx context.Context, k string, v interface{}) {
	GetSpan(ctx).DebugTag(k, v)
}

// Log add a log to a tracer context
func Log(ctx context.Context, args ...interface{}) {
	GetSpan(ctx).Log(args...)
}

// Logf add a log to a tracer context
func Logf(ctx context.Context, format string, args ...interface{}) {
	GetSpan(ctx).Logf(format, args...)
}

// DebugLog add a log to a tracer context if debug is enabled
func DebugLog(ctx context.Context, args ...interface{}) {
	GetSpan(ctx).DebugLog(args...)
}

// DebugLogf add a log to a tracer context if debug is enabled
func DebugLogf(ctx context.Context, format string, args ...interface{}) {
	GetSpan(ctx).DebugLogf(format, args...)
}
