FROM golang:1.21

RUN useradd -u 10001 scratchuser

COPY . /go/src/github.com/frzifus/lets-party

WORKDIR /go/src/github.com/frzifus/lets-party

RUN CGO_ENABLED=0 go build -v -o /lets-party

FROM scratch

COPY --from=0 /lets-party /lets-party
COPY --from=0 /etc/passwd /etc/passwd

USER scratchuser

EXPOSE 8080

CMD ["/lets-party", "-addr", "0.0.0.0:8080", "-log-level=info"]
