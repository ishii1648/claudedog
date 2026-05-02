package syncdb

import _ "embed"

//go:generate go run ./genhash

//go:embed schema.sql
var schemaSQL string
