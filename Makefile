ensure-fmt: fmt
	@git diff -s --exit-code *.go || (echo "Build failed: a go file is not formated correctly. Run 'make fmt' and update your PR." && exit 1)

fmt:
	go fmt ./...

vet:
	go vet ./...

build:
	CGO_ENABLED=0 go build -v ./...

test: fmt vet ensure-fmt
	CGO_ENABLED=0 go test -v ./... -failfast
