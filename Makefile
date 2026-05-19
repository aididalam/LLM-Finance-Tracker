.PHONY: dev dev-api dev-web build sync-web migrate-up migrate-down db-create test lint install-tools updatepi deployinpi fly-deploy

DB_USER     := $(shell grep -E '^DB_USER=' .env | cut -d '=' -f2)
DB_PASSWORD := $(shell grep -E '^DB_PASSWORD=' .env | cut -d '=' -f2)
DB_NAME     := $(shell grep -E '^DB_NAME=' .env | cut -d '=' -f2)

sync-web:
	cd web && npm ci && npm run build
	rsync -a --delete web/dist/ internal/web/files/

dev: 
	$(MAKE) -j2 dev-api dev-web

dev-api:
	$(shell go env GOPATH)/bin/air

dev-web:
	cd web && npm install && npm run dev

build: sync-web
	mkdir -p bin
	go build -ldflags="-s -w" -o bin/server ./cmd/server

migrate-up:
	go run ./cmd/migrate up

migrate-down:
	go run ./cmd/migrate down

db-create:
	mysql -u $(DB_USER) -p$(DB_PASSWORD) -e "CREATE DATABASE IF NOT EXISTS $(DB_NAME) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

test:
	go test ./...

lint:
	golangci-lint run

install-tools:
	go install github.com/air-verse/air@latest

updatepi: sync-web migrate-up
	mkdir -p bin
	go build -ldflags="-s -w" -o bin/server ./cmd/server
	sudo systemctl restart llmexpensetracker
	@echo "Deployed and restarted."

deployinpi: sync-web
	sudo systemctl stop llmexpensetracker || true
	mysql -u $(DB_USER) -p$(DB_PASSWORD) -e "DROP DATABASE IF EXISTS $(DB_NAME); CREATE DATABASE $(DB_NAME) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
	$(MAKE) migrate-up
	mkdir -p bin
	go build -ldflags="-s -w" -o bin/server ./cmd/server
	sudo systemctl start llmexpensetracker
	@echo "Fresh deploy complete."


