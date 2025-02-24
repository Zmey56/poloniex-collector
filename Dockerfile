# Dockerfile
FROM --platform=$BUILDPLATFORM golang:1.23.4-alpine AS builder

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

# Позволим Go самому определить архитектуру
RUN CGO_ENABLED=1 go build -o /app/bin/collector cmd/collector/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/bin/collector .
COPY internal/config/config.go ./config/

CMD ["./collector"]