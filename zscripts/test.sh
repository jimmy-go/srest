#!/bin/sh
cd $GOPATH/src/github.com/jimmy-go/srest

if [ "$1" == "bench" ]; then
    go test -race -bench=. -test.run=_NONE_
    exit;
fi

if [ "$1" == "html" ]; then
    go test -cover -coverprofile=coverage.out
    go tool cover -html=coverage.out
    exit;
fi

if [ "$1" == "allocs" ]; then
    TRACE_LOG=$GOBIN/srest.test.trace.log
    echo "generating: $TRACE_LOG"
    #go test -c -o $GOBIN/srest.test
    go test -race -c -o $GOBIN/srest.test
    GODEBUG=allocfreetrace=1 $GOBIN/stress -p=100 $GOBIN/srest.test -test.run=NONE -test.bench=$2 \
        -test.benchtime=5ms 2>$TRACE_LOG
    echo "view: $TRACE_LOG"
    exit;
fi

if [ "$1" == "stress" ]; then
    go test -c -o $GOBIN/srest.test
    $GOBIN/stress -p=10 $GOBIN/srest.test -test.run=. -test.bench=.
fi

#GODEBUG=allocfreetrace=0.1 go test -race -cover -test.run=$1 -bench=.
go test -race -cover -test.run=$1 -bench=.
