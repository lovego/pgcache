package logger

import (
	"context"
	"time"

	"github.com/lovego/tracer"
)

func (l *Logger) Record(
	debug bool, workFunc func(ctx context.Context) error,
	recoverFunc func(), fieldsFunc func(*Fields),
) {
	span := &tracer.Span{At: time.Now()}
	if debug {
		span.SetDebug(true)
	}
	var err error
	defer func() {
		panicErr := recover()
		if panicErr != nil && recoverFunc != nil {
			recoverFunc()
		}

		f := l.WithSpan(span)
		if fieldsFunc != nil {
			fieldsFunc(f)
		}

		if panicErr != nil {
			f.output(Recover, panicErr, f.data)
		} else if err != nil {
			f.output(Error, err, f.data)
		} else {
			f.output(Info, nil, f.data)
		}
	}()
	err = workFunc(tracer.Context(context.Background(), span))
}

func (l *Logger) WithSpan(span *tracer.Span) *Fields {
	span.Finish()
	f := l.With("at", span.At).With("duration", span.Duration)
	if len(span.Children) > 0 {
		f = f.With("children", span.Children)
	}
	if len(span.Tags) > 0 {
		f = f.With("tags", span.Tags)
	}
	return f
}
