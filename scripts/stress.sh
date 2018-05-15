#!/bin/bash
## DeGOps: 0.0.4
set -o errexit
set -o nounset

go get -u golang.org/x/tools/cmd/stress

TMP=tmp_log
mkdir -p $TMP

TRACE_LOG=$TMP/srest.test.trace.log
echo "generating: $TRACE_LOG"
go test -v -race -c -o $TMP/srest.test

GODEBUG=allocfreetrace=1 $GOBIN/stress -p=100 $TMP/srest.test -test.run=NONE -test.bench=. \
    -test.benchtime=5ms 2>$TRACE_LOG
echo "view: $TRACE_LOG"

