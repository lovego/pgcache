# logger
a logger that integrate with alarm.

[![Build Status](https://travis-ci.org/lovego/logger.svg?branch=new_json)](https://travis-ci.org/lovego/logger)
[![Coverage Status](https://coveralls.io/repos/github/lovego/logger/badge.svg?branch=new_json)](https://coveralls.io/github/lovego/logger?branch=new_json)
[![Go Report Card](https://goreportcard.com/badge/github.com/lovego/logger)](https://goreportcard.com/report/github.com/lovego/logger)
[![GoDoc](https://godoc.org/github.com/lovego/logger?status.svg)](https://godoc.org/github.com/lovego/logger)

## Install
`$ go get github.com/lovego/logger`

## Usage
```go
logger := New(os.Stdout)

logger.SetLevel(Debug)
logger.Debug("the ", "message")
logger.Debugf("this is %s", "test")

logger.Info("the ", "message")
logger.Infof("this is a %s", "test")

logger.Error("err")
logger.Errorf("test %s", "errorf")

logger.Panic("panic !!")
logger.Panicf("test %s", "panicf")

logger.Fatal("fatal !!")
logger.Fatalf("test %s", "fatalf")

defer logger.Recover()

logger.Record(os.Getenv("debug") != "", func(ctx context.Context) error {
  // work to do goes here
  return nil
}, nil, func(f *Fields) {
  f.With("key1", "value1").With("key2", "value2")
})
```

## Documentation
  [https://godoc.org/github.com/lovego/logger](https://godoc.org/github.com/lovego/logger)
