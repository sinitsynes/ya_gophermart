package repository_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/sinitsynes/ya_gophermart/internal/testutil"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

var PostgresContainer *postgres.PostgresContainer

func TestMain(m *testing.M) {
	ctx := context.Background()
	newCtr, err := testutil.NewPostgresContainer(ctx, "gophermart")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	PostgresContainer = newCtr
	code := m.Run()
	PostgresContainer.Terminate(ctx)
	os.Exit(code)
}
