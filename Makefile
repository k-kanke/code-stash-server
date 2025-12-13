.PHONY: migrate wire seed

migrate:
	docker compose run --rm migrate

wire:
	wire ./cmd/api

seed:
	go run ./cmd/seed
