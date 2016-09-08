#|/bin/sh

PKG_DIR=$GOPATH/src/github.com/jimmy-go/srest/examples/simple
cd $PKG_DIR/cmd/server

go build -race -o $GOBIN/simplerest

$GOBIN/simplerest \
    -port=2121 \
    -templates=$PKG_DIR/views \
    -static=$PKG_DIR/static \
    -db="host=192.168.2.16 dbname=lostsdb port=5432 user=postgres password=xx123456"
