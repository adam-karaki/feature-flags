.PHONY: run test vet build fmt infra-up infra-down

run:
	go run ./cmd/server

test:
	go test ./...

vet:
	go vet ./...

build:
	go build -o bin/flags-server ./cmd/server

fmt:
	gofmt -w $$(find . -name '*.go' -not -path './vendor/*')

infra-up:
	docker compose up -d

infra-down:
	docker compose down
