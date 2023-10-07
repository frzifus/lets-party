build:
	CGO_ENABLED=0 go build -v ./...

test:
	CGO_ENABLED=0 go test -v ./... -failfast
