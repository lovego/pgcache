package errs

import (
	"bytes"
	"fmt"
	"runtime"
)

func Stack(skip int) string {
	buf := new(bytes.Buffer)

	callers := make([]uintptr, 32)
	n := runtime.Callers(skip, callers)
	frames := runtime.CallersFrames(callers[:n])
	for {
		if f, ok := frames.Next(); ok {
			fmt.Fprintf(buf, "%s\n\t%s:%d (0x%x)\n", f.Function, f.File, f.Line, f.PC)
		} else {
			break
		}
	}
	return buf.String()
}

func WithStack(err error) string {
	if e, ok := err.(interface {
		Stack() string
	}); ok {
		return err.Error() + "\n" + e.Stack()
	}
	return err.Error()
}
