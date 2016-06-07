#|/bin/sh
cd $GOPATH/src/github.com/jimmy-go/srest/examples/simple

go build -race -o $GOBIN/simplerest

gcvis $GOBIN/simplerest \
-port=2121 \
-templates=$(pwd)/views \
-static=$(pwd)/static \
-workers=100 \
-queue=100 \
-db="host=192.168.2.10 dbname=lostsdb port=5432 user=postgres password=xx123456"
