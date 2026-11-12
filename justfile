lint:
    golangci-lint run --fix

run:
    docker compose -f deployments/gophermart/docker-compose.yaml up api --build

test:
    go test -coverpkg=./... -coverprofile=coverage.out ./... && go tool cover -func=coverage.out | tail -n 1

generate-sqlc:
    go tool sqlc generate

migrate-create name:
    export GOOSE_MIGRATION_DIR=./migrations/gophermart && \
    go tool goose create {{name}} sql
