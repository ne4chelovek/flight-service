include .env

LOCAL_BIN := $(CURDIR)/bin

install-deps:
	GOBIN=$(LOCAL_BIN) go install github.com/pressly/goose/v3/cmd/goose@v3.15.1


migrate-up:
	goose -dir migrations/ postgres "$(DB_URL)" up

migrate-down:
	goose -dir migrations/ postgres "$(DB_URL)" down

migrate-status:
	goose -dir migrations/ postgres "$(DB_URL)" status