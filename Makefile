
lint:
	golangci-lint run

test:
	@if [ -f .env ]; then \
		export $$(cat .env | xargs) && go test -v -coverprofile=coverage.out ./...; \
	else \
		go test -v -coverprofile=coverage.out ./...; \
	fi
	go tool cover -func=coverage.out
