
help:
	@echo "Available commands:"
	@echo "make install			Install dependencies."
	@echo "make test			Run tests."
	@echo "make bench			Run benchmarks."
	@echo "make coverage			Show coverage in html."
	@echo "make stress			Stress test."
	@echo "make clean			Clean build files."

install:
	@echo "Make: Install"
	glide up

.PHONY: test
test:
	@echo "Make: Test"
	go test -v -race -cover

bench:
	@echo "Make: Benchmarking"
	./zscripts/test.sh bench

coverage:
	@echo "Make: Coverage"
	./zscripts/test.sh html

stress:
	@echo "Make: Stress"
	./zscripts/test.sh allocs

clean:
	@echo "Make: Clean"
	rm -rf vendor
	rm -rf temp
	rm -rf _tmp_views
	touch coverage.out
	rm coverage.out
