#|/bin/sh
cd $GOPATH/src/github.com/jimmy-go/srest/examples/simple

go build -race -o $GOBIN/simplerest

$GOBIN/simplerest \
-port=2121 \
-templates=$(pwd) \
-db=$REST_DB
