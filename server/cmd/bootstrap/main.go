package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"

	usecases "milpa/aplication/use-cases"
	"milpa/infrastructure/adapters/secondary/auth"
	repo "milpa/infrastructure/adapters/secondary/repository"
	timepkg "milpa/infrastructure/adapters/secondary/time"
	"milpa/infrastructure/config"
	"milpa/infrastructure/database"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, using system env")
	}

	cfg := config.Load()

	email := os.Getenv("ADMIN_EMAIL")
	password := os.Getenv("ADMIN_PASSWORD")
	firstName := envOrDefault("ADMIN_FIRST_NAME", "Admin")
	lastName := envOrDefault("ADMIN_LAST_NAME", "Milpa")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := database.CreatePool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer database.CloseConnection(pool)

	database.MakeMigrations(context.Background(), cfg.DatabaseURL)

	hasher := auth.NewBcryptHasher(0)
	clock := timepkg.NewClock()
	userRepo := repo.NewUserRepository(pool)
	bootstrapUC := usecases.NewBootstrapAdminUseCase(userRepo, hasher, clock)

	result, err := bootstrapUC.EnsureAdmin(ctx, email, password, firstName, lastName)
	if err != nil {
		log.Fatalf("bootstrap: failed to ensure admin: %v", err)
	}

	switch {
	case result.Created:
		log.Printf("bootstrap: admin created (email=%s)", result.Email)
	case result.Promoted:
		log.Printf("bootstrap: existing user promoted to admin (email=%s)", result.Email)
	default:
		log.Printf("bootstrap: user is already admin, nothing to do (email=%s)", result.Email)
	}
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
