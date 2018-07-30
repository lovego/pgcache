# tracer
a trace helper for golang.

[![Build Status](https://travis-ci.org/lovego/tracer.svg?branch=master)](https://travis-ci.org/lovego/tracer)
[![Coverage Status](https://coveralls.io/repos/github/lovego/tracer/badge.svg?branch=master)](https://coveralls.io/github/lovego/tracer?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/lovego/tracer)](https://goreportcard.com/report/github.com/lovego/tracer)
[![GoDoc](https://godoc.org/github.com/lovego/tracer?status.svg)](https://godoc.org/github.com/lovego/tracer)

## Install
`$ go get github.com/lovego/tracer`

## Usage
```go
func main() {
  span := &tracer.Span{ At: time.Now() }
  defer span.Finish()
  ctx := tracer.Context(context.Background(), span)
  work(ctx)
}

func work(ctx context.Context) {
  tracer.Tag(ctx, "key", "value")
}
```

## Documentation
[https://godoc.org/github.com/lovego/tracer](https://godoc.org/github.com/lovego/tracer)
