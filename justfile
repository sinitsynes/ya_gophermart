set dotenv-load

lint:
    golangci-lint run --fix

run:
    docker compose -f deployments/gophermart/docker-compose.yaml up api --build

test:
    go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out | tail -n 1

generate-sqlc:
    go tool sqlc generate

migrate-create name:
    export GOOSE_MIGRATION_DIR=./migrations/gophermart && \
    go tool goose create {{name}} sql

migrate-up:
    export GOOSE_MIGRATION_DIR=./migrations/gophermart && \
    export GOOSE_DBSTRING=$DATABASE_DSN && \
    export GOOSE_DRIVER=postgres && \
    go tool goose up
