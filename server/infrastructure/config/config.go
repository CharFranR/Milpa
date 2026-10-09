package config

import (
	"fmt"
	"milpa/aplication/dto"
	"net/url"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	ESClient    dto.ESClient
	JWTSecret   string
	ServerPort  string
	Redis       Redis
}

type Redis struct {
	Addr     string
	Password string
	TLS      bool
}

const DefaultESIndex = "milpa-offerings"

const DefaultServerPort = "8080"

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
		Redis:       loadRedis(),
	}
}

func loadRedis() Redis {
	raw := os.Getenv("REDIS_URL")
	if raw == "" {
		return Redis{
			Addr:     os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
			Password: os.Getenv("REDIS_PASSWORD"),
		}
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return Redis{Addr: raw, Password: os.Getenv("REDIS_PASSWORD")}
	}

	addr := parsed.Host
	if addr == "" {
		addr = raw
	}

	password, _ := parsed.User.Password()

	return Redis{Addr: addr, Password: password, TLS: parsed.Scheme == "rediss"}
}
