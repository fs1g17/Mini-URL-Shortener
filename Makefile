export GOOSE_DRIVER := postgres
export GOOSE_DBSTRING := postgresql://postgres:postgres@localhost:5432/postgres
export GOOSE_MIGRATION_DIR := $(CURDIR)/migrations

up:
	docker compose up -d --wait
	goose up

down:
	goose reset
	docker compose down

test:
	go test ./... -count=1
