# GO
build:
	go build -o bin/api ./cmd/api

run:
	go run ./cmd/api

test:
	go test ./... -v -race -count=1

lint:
	go vet ./...

clean:
	rm -rf bin/

# Docker

up:
	docker compose up -d

down:
	docker compose down

down-clean:
	docker compose down -v

restart-api:
	docker compose restart api

seed:
	docker compose --profile tools build seed
	docker compose --profile tools run --rm seed

bootstrap:
	docker compose --profile tools build bootstrap
	docker compose --profile tools run --rm bootstrap

logs-api:
	docker compose logs -f api

logs-db:
	docker compose logs -f db

.PHONY: build run test lint clean up down down-clean restart-api logs-api logs-db seed bootstrap
