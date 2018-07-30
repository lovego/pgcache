package logger

import (
	"fmt"
	"os"
)

type Fields struct {
	*Logger
	data map[string]interface{}
}

// don't use (level, at, msg, stack, duration) as key, they will be overwritten.
func (f *Fields) With(key string, value interface{}) *Fields {
	f.data[key] = value
	return f
}

func (f *Fields) Debug(args ...interface{}) bool {
	if len(args) > 0 && f.level >= Debug {
		f.output(Debug, fmt.Sprint(args...), f.data)
	}
	return f.level >= Debug
}

func (f *Fields) Debugf(format string, args ...interface{}) {
	if f.level >= Debug {
		f.output(Debug, fmt.Sprintf(format, args...), f.data)
	}
}

func (f *Fields) Info(args ...interface{}) bool {
	if len(args) > 0 && f.level >= Info {
		f.output(Info, fmt.Sprint(args...), f.data)
	}
	return f.level >= Info
}

func (f *Fields) Infof(format string, args ...interface{}) {
	if f.level >= Info {
		f.output(Info, fmt.Sprintf(format, args...), f.data)
	}
}

func (f *Fields) Error(args ...interface{}) {
	f.output(Error, fmt.Sprint(args...), f.data)
}

func (f *Fields) Errorf(format string, args ...interface{}) {
	f.output(Error, fmt.Sprintf(format, args...), f.data)
}

func (f *Fields) Recover() {
	if err := recover(); err != nil {
		f.output(Recover, fmt.Sprint(err), f.data)
	}
}

func (f *Fields) Panic(args ...interface{}) {
	msg := fmt.Sprint(args...)
	f.output(Panic, fmt.Sprint(args...), f.data)
	panic(msg)
}

func (f *Fields) Panicf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	f.output(Panic, msg, f.data)
	panic(msg)
}

func (f *Fields) Fatal(args ...interface{}) {
	f.output(Fatal, fmt.Sprint(args...), f.data)
	os.Exit(1)
}

func (f *Fields) Fatalf(format string, args ...interface{}) {
	f.output(Fatal, fmt.Sprintf(format, args...), f.data)
	os.Exit(1)
}
