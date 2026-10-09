package config_test

import (
	"testing"

	"milpa/infrastructure/config"
)

// The connection info a managed Redis hands out is a single URL, and it is the
// only place the password and whether TLS is needed are stated. Cutting the URL
// at the first "@" (the old helper) kept the address but silently dropped both,
// so a rediss:// URL connected in the clear to a server that only speaks TLS.
func TestLoadRedisFromInternalURL(t *testing.T) {
	t.Setenv("REDIS_URL", "redis://red-abc123:6379")
	t.Setenv("REDIS_HOST", "ignored")
	t.Setenv("REDIS_PORT", "1111")
	t.Setenv("REDIS_PASSWORD", "ignored")

	cfg := config.Load()

	if cfg.Redis.Addr != "red-abc123:6379" {
		t.Errorf("addr = %q, want %q", cfg.Redis.Addr, "red-abc123:6379")
	}
	if cfg.Redis.Password != "" {
		t.Errorf("password = %q, want empty for an internal URL", cfg.Redis.Password)
	}
	if cfg.Redis.TLS {
		t.Error("TLS = true, want false for a redis:// URL")
	}
}

func TestLoadRedisFromExternalURL(t *testing.T) {
	t.Setenv("REDIS_URL", "rediss://default:s3cr3t@red-abc123:6379")
	t.Setenv("REDIS_PASSWORD", "ignored")

	cfg := config.Load()

	if cfg.Redis.Addr != "red-abc123:6379" {
		t.Errorf("addr = %q, want %q", cfg.Redis.Addr, "red-abc123:6379")
	}
	if cfg.Redis.Password != "s3cr3t" {
		t.Errorf("password = %q, want %q", cfg.Redis.Password, "s3cr3t")
	}
	if !cfg.Redis.TLS {
		t.Error("TLS = false, want true for a rediss:// URL")
	}
}

func TestLoadRedisPrefersURLOverDiscreteVariables(t *testing.T) {
	t.Setenv("REDIS_URL", "redis://cache:6380")
	t.Setenv("REDIS_HOST", "otro")
	t.Setenv("REDIS_PORT", "1111")
	t.Setenv("REDIS_PASSWORD", "clave")

	cfg := config.Load()

	if cfg.Redis.Addr != "cache:6380" {
		t.Errorf("addr = %q, want the one from REDIS_URL", cfg.Redis.Addr)
	}
	if cfg.Redis.Password != "" {
		t.Errorf("password = %q, want empty: the URL carries none", cfg.Redis.Password)
	}
}

func TestLoadRedisWithoutURL(t *testing.T) {
	t.Setenv("REDIS_URL", "")
	t.Setenv("REDIS_HOST", "localhost")
	t.Setenv("REDIS_PORT", "6379")
	t.Setenv("REDIS_PASSWORD", "clave")

	cfg := config.Load()

	if cfg.Redis.Addr != "localhost:6379" {
		t.Errorf("addr = %q, want %q", cfg.Redis.Addr, "localhost:6379")
	}
	if cfg.Redis.Password != "clave" {
		t.Errorf("password = %q, want %q", cfg.Redis.Password, "clave")
	}
	if cfg.Redis.TLS {
		t.Error("TLS = true, want false when nothing asks for it")
	}
}

func TestLoadRedisKeepsUnparsableURLAsAddress(t *testing.T) {
	t.Setenv("REDIS_URL", "cache:6379")
	t.Setenv("REDIS_PASSWORD", "clave")

	cfg := config.Load()

	if cfg.Redis.Addr != "cache:6379" {
		t.Errorf("addr = %q, want the value used as-is", cfg.Redis.Addr)
	}
}