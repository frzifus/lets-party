FROM golang:1.22

COPY . /go/src/github.com/quixsi/core

WORKDIR /go/src/github.com/quixsi/core

RUN CGO_ENABLED=0 go build -v -o /lets-party cmd/server/main.go

FROM scratch

COPY --from=0 /lets-party /lets-party

EXPOSE 8080

CMD ["/lets-party", "-addr", "0.0.0.0:8080", "-log-level=info"]
