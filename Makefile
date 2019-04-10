.PHONY: bench
bench:
	go test -bench=. ./pkg/cache/


.PHONY: coverage
coverage:
	go test -coverprofile cover.out ./pkg/cache/ && \
	go tool cover -func=cover.out