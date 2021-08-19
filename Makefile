
.PHONY: deps
deps:
	@go mod download
	@go mod vendor
	@go mod tidy

.PHONY: build
build:
	@go build -o ./bin/bot ./

.PHONY: clean
clean:
	@rm -fv ./bin/*

.PHONY: generate
generate: tools
	@export PATH=$(shell pwd)/bin:$(PATH); go generate ./...

.PHONY: tools
tools: deps
	@go install github.com/gojuno/minimock/v3/cmd/minimock

.PHONY: test
test:
	@go test ./...

.PHONY: lint
lint:
	@golangci-lint run
