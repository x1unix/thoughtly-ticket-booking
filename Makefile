export ENV_FILE=$(CURDIR)/local.env

.PHONY: help
help:
	@echo 'Usage: make [run|test|migrate-up|migrate-new]'

.PHONY: run
run:
	@go run ./cmd/server

.PHONY: test
test:
	@go test -v -count 1 $(TEST_CMD) ./tests

.PHONY: migrate-up
migrate-up:
	@go run ./cmd/migrate up

.PHONY: migrate-new
migrate-new:
	@[ -z "$(MIGRATION_NAME)" ] && echo "Missing MIGRATION_NAME" || go run ./cmd/migrate create $(MIGRATION_NAME) sql

.PHONY: migrate-version
migrate-version:
	@go run ./cmd/migrate version

