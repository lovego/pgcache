package tracer

import (
	"context"
	"fmt"
	"time"
)

// Span for tracing
type Span struct {
	Name     string                 `json:"name,omitempty"`
	At       time.Time              `json:"at"`
	Duration float64                `json:"duration"` // milliseconds
	Children []*Span                `json:"children,omitempty"`
	Tags     map[string]interface{} `json:"tags,omitempty"`
	Logs     []string               `json:"logs,omitempty"`
	debug    bool
}

// StartSpan start a new span on the given context
func StartSpan(ctx context.Context, name string) *Span {
	if ctx != nil {
		if value := ctx.Value(spanKey); value != nil {
			if parent, ok := value.(*Span); ok && parent != nil {
				s := &Span{Name: name, At: time.Now()}
				parent.Children = append(parent.Children, s)
				return s
			}
		}
	}
	return nil
}

// Finish finish a span
func (s *Span) Finish() {
	if s != nil {
		s.Duration = float64(time.Since(s.At)) / float64(time.Millisecond)
	}
}

// GetDebug returns if debug is enabled
func (s *Span) GetDebug() bool {
	if s != nil {
		return s.debug
	}
	return false
}

// SetDebug set if debugging tag should be traced.
func (s *Span) SetDebug(b bool) *Span {
	if s != nil {
		s.debug = b
	}
	return s
}

// Tag add a tag to a span
func (s *Span) Tag(k string, v interface{}) *Span {
	if s == nil {
		return s
	}
	if s.Tags == nil {
		s.Tags = make(map[string]interface{})
	}
	s.Tags[k] = v
	return s
}

// DebugTag add a tag to a span if debug is enabled
func (s *Span) DebugTag(k string, v interface{}) *Span {
	if s == nil || !s.debug {
		return s
	}
	if s.Tags == nil {
		s.Tags = make(map[string]interface{})
	}
	s.Tags[k] = v
	return s
}

// Log add a log to a span using fmt.Sprint
func (s *Span) Log(args ...interface{}) *Span {
	if s == nil {
		return s
	}
	s.Logs = append(s.Logs, fmt.Sprint(args...))
	return s
}

// Logf add a log to a span using fmt.Sprintf
func (s *Span) Logf(format string, args ...interface{}) *Span {
	if s == nil {
		return s
	}
	s.Logs = append(s.Logs, fmt.Sprintf(format, args...))
	return s
}

// DebugLog add a log to a span using fmt.Sprint if debug is enabled
func (s *Span) DebugLog(args ...interface{}) *Span {
	if s == nil || !s.debug {
		return s
	}
	s.Logs = append(s.Logs, fmt.Sprint(args...))
	return s
}

// DebugLogf add a log to a span using fmt.Sprintf if debug is enabled
func (s *Span) DebugLogf(format string, args ...interface{}) *Span {
	if s == nil || !s.debug {
		return s
	}
	s.Logs = append(s.Logs, fmt.Sprintf(format, args...))
	return s
}
