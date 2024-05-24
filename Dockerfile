FROM golang:1.21-alpine AS builder

COPY . /app

WORKDIR /app

RUN go build -ldflags "-s -w" -o tpser main.go

FROM alpine

COPY --from=builder /app/tpser /usr/local/bin/tpser

ENTRYPOINT ["/usr/local/bin/tpser"]