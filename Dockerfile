FROM golang:1.19 AS builder

WORKDIR /go/src/tbcpusher
COPY . .
RUN CGO_ENABLED=0 go build -o /go/bin/tbcpusher

FROM gcr.io/distroless/static-debian11
COPY --from=builder /go/bin/tbcpusher /

EXPOSE 8000

CMD ["/tbcpusher"]