#!/bin/bash
## DeGOps: 0.0.4
set -o errexit
set -o nounset

go test -v -race -bench=. -test.run=_NONE_
