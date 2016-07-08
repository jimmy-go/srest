#!/bin/sh
cd $GOPATH/src/github.com/jimmy-go/srest

if [ "$1" == "bench" ]; then
    go test -race -bench=.
    exit;
fi

go test -cover -coverprofile=coverage.out

if [ "$1" == "html" ]; then
    go tool cover -html=coverage.out
fi
