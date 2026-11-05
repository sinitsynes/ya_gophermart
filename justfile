lint:
    golangci-lint run --fix

run:
    docker compose up api --build

test:
    go test ./...

generate-sqlc:
    go tool sqlc generate

migrate-create name:
    export GOOSE_MIGRATION_DIR=./migrations/gophermart && \
    go tool goose create {{name}} sql
