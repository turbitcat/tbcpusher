FROM golang:1.19 AS builder

WORKDIR /go/src/telegram
COPY . .
RUN CGO_ENABLED=0 go build -o /go/bin/telegram

FROM gcr.io/distroless/static-debian11
COPY --from=builder /go/bin/telegram /

EXPOSE 8000

CMD ["/telegram"]