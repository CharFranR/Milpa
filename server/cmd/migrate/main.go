package main

import (
	"context"

	"milpa/infrastructure/config"
	"milpa/infrastructure/database"
)

func main() {
	cfg := config.Load()
	database.MakeMigrations(context.Background(), cfg.DatabaseURL)
}
