package migrations

import (
	"embed"
)

//go:embed gophermart/*.sql
var EmbedMigrations embed.FS
