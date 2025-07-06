
lint:
	go vet ./...

test:
	@if [ -f .env ]; then \
		export $$(cat .env | xargs) && go test -v ./...; \
	else \
		echo "Warning: .env file not found, running tests without environment variables"; \
		go test -v ./...; \
	fi