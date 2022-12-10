GOLANGCI_LINT ?= go run github.com/golangci/golangci-lint/cmd/golangci-lint@latest
MOCKERY ?= go run github.com/vektra/mockery/v2@latest

.PHONY: build
build: generate
	go build -o . ./cmd/...

.PHONY: generate
generate:
	go generate ./...

.PHONY: check
check: static-analysis test

.PHONY: static-analysis
static-analysis: generate mocks
	$(GOLANGCI_LINT) run

.PHONY: lint
lint: static-analysis

.PHONY: fix
fix: generate mocks
	$(GOLANGCI_LINT) run --fix

.PHONY: test
test: test-unit

.PHONY: test-unit
test-unit: generate mocks
	go test --timeout 1m --shuffle on ./...
	go test --timeout 1m --shuffle on --race --short ./...
	go test --timeout 1m --shuffle on --count 100 --short ./...

mocks: Makefile $(shell find . -type f -name '*.go' -not -name '*_test.go' -not -path './mocks/*')
	$(MOCKERY) --with-expecter --all --keeptree
	@touch mocks

.PHONY: clean
clean:
	rm -rf mocks
