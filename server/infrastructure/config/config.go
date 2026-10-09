package config

import (
	"fmt"
	"net/url"
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
	Redis       Redis
}

// Redis is the cache connection, resolved from REDIS_URL when it is set and
// from the discrete REDIS_* variables otherwise.
//
// Addr and Password are handed to the client as they are, and TLS says whether
// the connection has to be encrypted. Render's internal Key Value URL is
// redis://host:6379 with no credentials and no TLS, while its external one is
// rediss:// with both; only the first used to survive the parsing, which is why
// the URL is read here instead of being cut at the first "@" by each command.
type Redis struct {
	Addr     string
	Password string
	TLS      bool
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
		Redis:       loadRedis(),
	}
}

// loadRedis resolves the cache connection, preferring REDIS_URL over the
// discrete variables because that is how every managed Redis hands out its
// connection info, credentials and TLS included.
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
		// A value that does not parse at all is used as the address: the caller
		// has no way to recover from an error here, and a bare "host:port" is a
		// reasonable thing to have been handed.
		return Redis{Addr: raw, Password: os.Getenv("REDIS_PASSWORD")}
	}

	// Without "//" net/url reads "cache:6379" as scheme "cache" with the port
	// in Opaque and an empty Host, so the address has to come from the raw value
	// unless the URL actually carries an authority.
	addr := parsed.Host
	if addr == "" {
		addr = raw
	}

	password, _ := parsed.User.Password()

	// "rediss" is TLS for Redis. Any other scheme is taken at face value: the
	// client is told to encrypt only when the URL says so.
	return Redis{Addr: addr, Password: password, TLS: parsed.Scheme == "rediss"}
}
