# Build stage — Go server with CGO (SQLite requires it)
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 go build -ldflags "-s -w" -o bitrok-server ./server/cmd/bitrok-server

# Runtime stage — minimal image
FROM alpine:latest

RUN apk add --no-cache ca-certificates curl
RUN addgroup -S bitrok && adduser -S bitrok -G bitrok

WORKDIR /app
RUN mkdir -p /data && chown bitrok:bitrok /data

COPY --from=builder /app/bitrok-server .

USER bitrok
EXPOSE 8080

VOLUME ["/data"]
ENV BITROK_DB_PATH=/data/bitrok.db

ENTRYPOINT ["./bitrok-server"]