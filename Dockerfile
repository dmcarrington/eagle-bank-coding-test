FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /eagle-bank ./cmd/server

FROM alpine:3.20

RUN apk add --no-cache su-exec && \
    addgroup -S eagle && adduser -S eagle -G eagle && \
    mkdir -p /data

WORKDIR /app

COPY --from=builder /eagle-bank .
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

EXPOSE 8090

ENTRYPOINT ["/entrypoint.sh"]
