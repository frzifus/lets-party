include ./Makefile.Common

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
	$(GO_ENV)
	go build -v ./...

run: gofmt build
	$(LETS_PARTY)

gotest: 
	$(GO_ENV)
	go test -v ./... -failfast

localtest: gofmt govet check-fmt
	$(GO_ENV)
	go test -v ./... -failfast

gomoddownload:
	go mod download -x

install-gotools: $(TOOLS_DIR)
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(TOOLS_DIR) $(GOLINT_VERSION) 

golint:
	$(LINT) run --verbose --allow-parallel-runners --timeout=10m 

gotidy:
	go mod tidy -compat=$(GO_VERSION)
