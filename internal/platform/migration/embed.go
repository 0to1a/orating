// Package migration embeds SQL files for golang-migrate. Migrations run on
// startup from cmd/server/main.go and are idempotent.
//
// Create a new pair via `make migrate-create name=add_xxx` (uses migrate CLI
// for collision-free filenames).
package migration

import "embed"

//go:embed sql/*.sql
var FS embed.FS
