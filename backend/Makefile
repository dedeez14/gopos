# Perintah harian backend. `make <target>`.

.PHONY: run build tidy swag test

run: ## jalankan server dev (env dari .env)
	set -a; . ./.env; set +a; go run ./cmd/api

build: ## binari produksi → bin/tuleh-server
	CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/tuleh-server ./cmd/api

tidy: ## rapikan go.mod/go.sum
	go mod tidy

swag: ## regenerasi dokumentasi Swagger → docs/
	swag init -g cmd/api/main.go

test:
	go test ./...

sinkron: ## salin master data dari MOVERA (baca-saja, idempoten)
	set -a; . ./.env; set +a; go run ./cmd/sinkron
