package repository

import "embed"

//go:embed migrations
var Migrations embed.FS
