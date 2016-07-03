#!/bin/sh
cd $GOPATH/src/github.com/jimmy-go/srest

if [ "$1" == "bench" ]; then
    go test -race -bench=.
fi

if [ "$1" == "normal" ]; then
    go test -cover -coverprofile=coverage.out
fi

if [ "$1" == "html" ]; then
    go test -cover -coverprofile=coverage.out
    go tool cover -html=coverage.out
fi
