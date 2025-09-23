work_dir = $(shell pwd)
golangci_version = $(shell head -n 1 .golangci.yml | tr -d '\# ')

all: build

.PHONY: build
build:
	go build -o build/k6build ./cmd/k6build

.PHONY: format
format:
	gofmt -w -s .
	gofumpt -w .

# Running with -buildvcs=false works around the issue of `go list all` failing when git, which runs as root inside
# the container, refuses to operate on the disruptor source tree as it is not owned by the same user (root).
.PHONY: lint
lint: format
	docker run --rm -v $(work_dir):/src -w /src -e GOFLAGS=-buildvcs=false golangci/golangci-lint:$(golangci_version) golangci-lint run

.PHONY: test
test:
	go test -race  ./...

.PHONY: integration
integration:
	go test -tags integration -race  ./integration/...

.PHONY: readme
readme:
	go run ./tools/gendoc README.md
