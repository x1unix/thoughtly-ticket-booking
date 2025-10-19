.PHONY: help
help:
	@echo 'Usage: make [run|migrate-up|migrate-new]'

.PHONY: run
run:
	@go run ./cmd/server -e local.env

.PHONY: migrate-up
migrate-up:
	@go run ./cmd/migrate -e local.env up

.PHONY: migrate-new
migrate-new:
	@[ -z "$(MIGRATION_NAME)" ] && echo "Missing MIGRATION_NAME" || go run ./cmd/migrate -e local.env create $(MIGRATION_NAME) sql

.PHONY: migrate-version
migrate-version:
	@go run ./cmd/migrate -e local.env version

