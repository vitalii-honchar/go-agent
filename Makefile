
build: lint
	go build ./pkg/goagent/...

fmt:
	gofmt -w .

lint: fmt
	golangci-lint run

test:
	@if [ -f .env ]; then \
		export $$(cat .env | xargs) && go test -v -coverprofile=coverage.out ./pkg/...; \
	else \
		go test -v -coverprofile=coverage.out ./pkg/...; \
	fi
	go tool cover -func=coverage.out

.PHONY: build fmt lint test
