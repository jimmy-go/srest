#|/bin/sh
cd $GOPATH/src/github.com/jimmy-go/srest/examples/simple/stest

go build -o $GOBIN/ttstress

$GOBIN/ttstress \
-static=$(pwd)/../static \
-users=20 \
-host=http://localhost:2121
