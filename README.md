# pgnotify
Make postgresql notify you when "UPDATE", "INSERT", "DELETE" executed, and return expected columns' data to you. 

[![Build Status](https://travis-ci.org/lovego/pgnotify.svg?branch=master)](https://travis-ci.org/lovego/pgnotify)
[![Coverage Status](https://img.shields.io/coveralls/github/lovego/pgnotify/master.svg)](https://coveralls.io/github/lovego/pgnotify?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/lovego/pgnotify)](https://goreportcard.com/report/github.com/lovego/pgnotify)
[![GoDoc](https://godoc.org/github.com/lovego/pgnotify?status.svg)](https://godoc.org/github.com/lovego/pgnotify)

## Install
`$ go get github.com/lovego/pgnotify`

## Test
`PG_DATA_SOURCE="postgres://user:password@host:port/db?sslmode=disable" go test`

## Docs
[https://godoc.org/github.com/lovego/pgnotify](https://godoc.org/github.com/lovego/pgnotify)
