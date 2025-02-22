# Dockerfile
FROM --platform=linux/amd64 golang:1.23.4-alpine AS builder

WORKDIR /app

RUN apk update && \
    apk add --no-cache \
    gcc \
    musl-dev \
    make

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -o /app/bin/collector cmd/collector/main.go

FROM --platform=linux/amd64 alpine:latest

WORKDIR /app

COPY --from=builder /app/bin/collector .
COPY internal/config/config.go ./config/

CMD ["./collector"]