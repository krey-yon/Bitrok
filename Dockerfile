# Build stage — Go server with CGO (SQLite requires it)
FROM golang:1.26.5-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY pkg ./pkg
COPY server ./server
ARG VERSION=dev
RUN CGO_ENABLED=1 go build -ldflags "-s -w -X github.com/bitrok/bitrok/server/internal/api.Version=${VERSION}" -o bitrok-server ./server/cmd/bitrok-server

# Runtime stage — minimal image
FROM alpine:3.23

RUN apk add --no-cache ca-certificates curl sqlite
RUN addgroup -S bitrok && adduser -S bitrok -G bitrok

WORKDIR /app
RUN mkdir -p /data && chown bitrok:bitrok /data

COPY --from=builder /app/bitrok-server .

USER bitrok
EXPOSE 8080

VOLUME ["/data"]
ENV BITROK_DB_PATH=/data/bitrok.db

ENTRYPOINT ["./bitrok-server"]
