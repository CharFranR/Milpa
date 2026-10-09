package config_test

import (
	"testing"

	"milpa/infrastructure/config"
)

// TestLoadDefaultsESIndex is the regression guard for the wiring defect: the
// search adapter was handed the zero value of the index name, so every search
// call targeted an empty index.
func TestLoadDefaultsESIndex(t *testing.T) {
	t.Setenv("ESCLIENT_INDEX", "")

	cfg := config.Load()

	if cfg.ESClient.Index != config.DefaultESIndex {
		t.Errorf("index = %q, want the default %q", cfg.ESClient.Index, config.DefaultESIndex)
	}
	if cfg.ESClient.Index == "" {
		t.Error("index is empty, want a resolvable name for a fresh checkout")
	}
}

func TestLoadReadsESIndexFromEnvironment(t *testing.T) {
	t.Setenv("ESCLIENT_INDEX", "custom-offerings")

	cfg := config.Load()

	if cfg.ESClient.Index != "custom-offerings" {
		t.Errorf("index = %q, want %q", cfg.ESClient.Index, "custom-offerings")
	}
}

func TestLoadDefaultsServerPort(t *testing.T) {
	t.Setenv("SERVER_PORT", "")

	cfg := config.Load()

	if cfg.ServerPort != config.DefaultServerPort {
		t.Errorf("server port = %q, want the default %q", cfg.ServerPort, config.DefaultServerPort)
	}
}
