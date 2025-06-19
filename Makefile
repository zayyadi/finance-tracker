build:
	@go build -o bin/finance ./cmd/server/main.go

test:
	@go test -v ./...
	
run: build
	@./bin/finance
.PHONY: build test run
clean:
	@rm -rf bin
	@rm -rf coverage.out
	@rm -rf .coverage
.PHONY: clean
coverage:
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
.PHONY: coverage
lint:
	@go fmt ./...
	@go vet ./...
	@golangci-lint run
.PHONY: lint
docker-build:
	@docker build -t finance:latest .
.PHONY: docker-build
docker-run:
	@docker run -p 8080:8080 --rm finance:latest
.PHONY: docker-run
docker-clean:
	@docker rmi finance:latest

