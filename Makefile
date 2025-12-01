migrate:
	docker compose run --rm migrate

wire:
	wire ./cmd/api
