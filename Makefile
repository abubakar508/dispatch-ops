BINARY := dispatch-ops
PKG := ./...

.PHONY: dev install run-api run-web build build-api test test-cover vet fmt tidy docker-build docker-run clean

dev:
	npm run dev

install:
	npm install

run-api:
	go run ./cmd/server

run-web:
	npm --prefix frontend run dev

build:
	npm --prefix frontend run build
	CGO_ENABLED=1 go build -trimpath -ldflags="-s -w" -o bin/$(BINARY) ./cmd/server

build-api:
	CGO_ENABLED=1 go build -trimpath -ldflags="-s -w" -o bin/$(BINARY) ./cmd/server

test:
	go test $(PKG)

test-cover:
	go test -cover -coverprofile=coverage.out $(PKG)
	go tool cover -func=coverage.out | tail -1

vet:
	go vet $(PKG)

fmt:
	gofmt -w -s .

tidy:
	go mod tidy

docker-build:
	docker build -t $(BINARY):latest .

docker-run:
	docker run --rm -p 8080:8080 --env-file .env $(BINARY):latest

clean:
	rm -rf bin coverage.out frontend/.next
