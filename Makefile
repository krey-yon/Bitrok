.PHONY: build build-cli build-server build-web dev dev-server dev-cli test clean install-cli release

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -X github.com/bitrok/bitrok/server/internal/api.Version=$(VERSION)

build: build-cli build-server

build-cli:
	go build -o bin/bitrok ./cli/cmd/bitrok

build-server:
	go build -ldflags "$(LDFLAGS)" -o bin/bitrok-server ./server/cmd/bitrok-server

build-web:
	cd web && npm run build

dev-server:
	go run ./server/cmd/bitrok-server

dev-cli:
	go run ./cli/cmd/bitrok $(ARGS)

test:
	go test ./...

clean:
	rm -rf bin/ bitrok.db bitrok bitrok-server

docker-build:
	docker compose build

docker-up:
	docker compose up -d

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f bitrok-server

install-cli: build-cli
	cp bin/bitrok /usr/local/bin/bitrok

release:
	goreleaser release --clean
