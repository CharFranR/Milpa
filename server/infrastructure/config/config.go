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

func Load() *Config {
	godotenv.Load()

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
		Index:               os.Getenv("ESCLIENT_INDEX"),
	}

	return &Config{
		DatabaseURL: DatabaseURL,
		ESClient:    ESClient,
		JWTSecret:   os.Getenv("JWT_SECRET"),
		ServerPort:  os.Getenv("SERVER_PORT"),
	}
}
