package config

import (
	"fmt"
	"milpa/aplication/dto"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	ESClient    dto.ESClient
	JWTSecret   string
	ServerPort  string
}

// DefaultESIndex is the index searched when ESCLIENT_INDEX is not set.
//
// Without a default a fresh checkout boots with an empty index name, and every
// search call silently targets nothing. The name has to be resolvable here,
// where the environment is read, rather than at each call site.
const DefaultESIndex = "milpa-offerings"

// DefaultServerPort mirrors the fallback the HTTP server applies when
// SERVER_PORT is unset.
const DefaultServerPort = "8080"

// envOrDefault returns the environment value for key, or fallback when it is
// unset or blank.
func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func Load() *Config {
	_ = godotenv.Load()

	DatabaseURL := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_DB"),
		os.Getenv("DB_SSLMODE"),
	)

	MaxIdleConnsPerHost, _ := strconv.Atoi(os.Getenv("ESCLIENT_MAXID"))

	// will use err in a log func later (or never)

	ESClient := dto.ESClient{
		Username:            os.Getenv("ESCLIENT_USER"),
		Password:            os.Getenv("ESCLIENT_PASSWORD"),
		Endpoint1:           os.Getenv("ESCLIENT_ENDPOINT1"),
		Endpoint2:           os.Getenv("ESCLIENT_ENDPOINT2"),
		MaxIdleConnsPerHost: MaxIdleConnsPerHost,
		Index:               envOrDefault("ESCLIENT_INDEX", DefaultESIndex),
	}

	return &Config{
		DatabaseURL: DatabaseURL,
		ESClient:    ESClient,
		JWTSecret:   os.Getenv("JWT_SECRET"),
		ServerPort:  envOrDefault("SERVER_PORT", envOrDefault("PORT", DefaultServerPort)),
	}
}
