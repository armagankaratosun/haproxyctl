package configuration

import "testing"

const (
	kindGlobal   = "Global"
	kindDefaults = "Defaults"
)

func TestGlobalConfig_Defaults(t *testing.T) {
	cfg := GlobalConfig{
		APIVersion: "haproxyctl/v1",
		Kind:       kindGlobal,
		Daemon:     true,
		Nbproc:     2,
		Maxconn:    2000,
		Log:        "stdout format raw local0",
	}

	if cfg.Kind != kindGlobal {
		t.Fatalf("expected Kind Global, got %s", cfg.Kind)
	}
	if !cfg.Daemon {
		t.Fatalf("expected Daemon true")
	}
	if cfg.Nbproc != 2 {
		t.Fatalf("expected Nbproc 2, got %d", cfg.Nbproc)
	}
	if cfg.Maxconn != 2000 {
		t.Fatalf("expected Maxconn 2000, got %d", cfg.Maxconn)
	}
}

func TestDefaultsConfig_Defaults(t *testing.T) {
	cfg := DefaultsConfig{
		APIVersion:     "haproxyctl/v1",
		Kind:           kindDefaults,
		Mode:           "http",
		TimeoutClient:  "30s",
		TimeoutServer:  "30s",
		TimeoutConnect: "5s",
		Balance:        "roundrobin",
	}

	if cfg.Kind != kindDefaults {
		t.Fatalf("expected Kind Defaults, got %s", cfg.Kind)
	}
	if cfg.Mode != "http" {
		t.Fatalf("expected Mode http, got %s", cfg.Mode)
	}
	if cfg.TimeoutClient != "30s" || cfg.TimeoutServer != "30s" {
		t.Fatalf("unexpected timeout values: %+v", cfg)
	}
	if cfg.Balance != "roundrobin" {
		t.Fatalf("expected Balance roundrobin, got %s", cfg.Balance)
	}
}
