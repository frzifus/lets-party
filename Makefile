#Set directory paths
ROOT_DIR=$(shell git rev-parse --show-toplevel)
TOOLS_DIR=$(ROOT_DIR)/.tools

#Set tool-paths for easier access
LINT := $(TOOLS_DIR)/golangci-lint

LETS_PARTY := $(ROOT_DIR)/bin/lets-party

#RULES
$(TOOLS_DIR):
	mkdir -p $@

check-fmt: gofmt
	@git diff -s --exit-code *.go || (echo "Build failed: a go file is not formated correctly. Run 'make fmt' and update your PR." && exit 1)

gofmt:
	go fmt ./...

govet:
	go vet ./...

build:
	go build -v -o $(LETS_PARTY) $(ROOT_DIR)/cmd/server/main.go

compilecheck:
	CGO_ENABLED=0 go build -v ./...

run: gofmt build
	$(LETS_PARTY)

gotest: 
	CGO_ENABLED=0 go test -v ./... -failfast

localtest: gofmt govet check-fmt
	CGO_ENABLED=0 go test -v ./... -failfast

gomoddownload:
	go mod download -x

install-gotools: $(TOOLS_DIR)
	#install golangci-lint v1.57.2 (only for workflow)
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(TOOLS_DIR) v1.57.2 

golint:
	$(LINT) run --verbose --allow-parallel-runners --timeout=10m 

gotidy:
	go mod tidy -compat=1.21
