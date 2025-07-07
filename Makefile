
lint:
	golangci-lint run

test:
	@if [ -f .env ]; then \
		export $$(cat .env | xargs) && go test -v ./...; \
	else \
		go test -v ./...; \
	fi