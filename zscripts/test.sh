#!/bin/sh

TMP=temp

mkdir -p $TMP

if [ "$1" == "bench" ]; then
    go test -v -race -bench=. -test.run=_NONE_
    exit;
fi

if [ "$1" == "html" ]; then
    go test -v -cover -coverprofile=coverage.out
    go tool cover -html=coverage.out
    exit;
fi

if [ "$1" == "allocs" ]; then
    TRACE_LOG=$TMP/srest.test.trace.log
    echo "generating: $TRACE_LOG"
    go test -v -race -c -o $TMP/srest.test
    GODEBUG=allocfreetrace=1 $GOBIN/stress -p=100 $TMP/srest.test -test.run=NONE -test.bench=$2 \
        -test.benchtime=5ms 2>$TRACE_LOG
    echo "view: $TRACE_LOG"
    exit;
fi

if [ "$1" == "stress" ]; then
    go test -v -c -o $TMP/srest.test
    $GOBIN/stress -p=10 $TMP/srest.test -test.run=. -test.bench=.
fi
