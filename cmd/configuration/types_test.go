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
		Nbproc:     3,
		Maxconn:    4000,
		Log:        "stdout format raw local0",
	}

	if cfg.Kind != kindGlobal {
		t.Fatalf("expected Kind Global, got %s", cfg.Kind)
	}
	if !cfg.Daemon {
		t.Fatalf("expected Daemon true")
	}
	if cfg.Nbproc != 3 {
		t.Fatalf("expected Nbproc 3, got %d", cfg.Nbproc)
	}
	if cfg.Maxconn != 4000 {
		t.Fatalf("expected Maxconn 4000, got %d", cfg.Maxconn)
	}
}

func TestDefaultsConfig_Defaults(t *testing.T) {
	const timeout30s = "30s"

	cfg := DefaultsConfig{
		APIVersion:     "haproxyctl/v1",
		Kind:           kindDefaults,
		Mode:           "http",
		TimeoutClient:  timeout30s,
		TimeoutServer:  timeout30s,
		TimeoutConnect: "5s",
		Balance:        "roundrobin",
	}

	if cfg.Kind != kindDefaults {
		t.Fatalf("expected Kind Defaults, got %s", cfg.Kind)
	}
	if cfg.Mode != "http" {
		t.Fatalf("expected Mode http, got %s", cfg.Mode)
	}
	if cfg.TimeoutClient != timeout30s || cfg.TimeoutServer != timeout30s {
		t.Fatalf("unexpected timeout values: %+v", cfg)
	}
	if cfg.Balance != "roundrobin" {
		t.Fatalf("expected Balance roundrobin, got %s", cfg.Balance)
	}
}

func TestMapGlobalFromAPIAndIsEmpty(t *testing.T) {
	const timeout30s = "30s"

	// Empty map should produce an "empty" config.
	emptyCfg := mapGlobalFromAPI(map[string]interface{}{})
	if !emptyCfg.isEmpty() {
		t.Fatalf("expected empty GlobalConfig to be reported as empty")
	}

	// A populated map should map all fields and not be empty.
	input := map[string]interface{}{
		"daemon":            true,
		"nbproc":            float64(2),
		"maxconn":           float64(2000),
		"log":               "stdout format raw local0",
		"log_send_hostname": "myhost",
		"stats_socket":      "/var/run/haproxy.sock",
		"stats_timeout":     timeout30s,
		"spread_checks":     float64(5),
	}

	cfg := mapGlobalFromAPI(input)

	if cfg.APIVersion != "haproxyctl/v1" {
		t.Fatalf("expected APIVersion haproxyctl/v1, got %s", cfg.APIVersion)
	}
	if cfg.Kind != kindGlobal {
		t.Fatalf("expected Kind %s, got %s", kindGlobal, cfg.Kind)
	}
	if cfg.Daemon != true {
		t.Fatalf("expected Daemon true, got %v", cfg.Daemon)
	}
	if cfg.Nbproc != 2 {
		t.Fatalf("expected Nbproc 2, got %d", cfg.Nbproc)
	}
	if cfg.Maxconn != 2000 {
		t.Fatalf("expected Maxconn 2000, got %d", cfg.Maxconn)
	}
	if cfg.Log != "stdout format raw local0" {
		t.Fatalf("unexpected Log: %s", cfg.Log)
	}
	if cfg.LogSendHost != "myhost" {
		t.Fatalf("unexpected LogSendHost: %s", cfg.LogSendHost)
	}
	if cfg.StatsSocket != "/var/run/haproxy.sock" {
		t.Fatalf("unexpected StatsSocket: %s", cfg.StatsSocket)
	}
	if cfg.StatsTimeout != timeout30s {
		t.Fatalf("unexpected StatsTimeout: %s", cfg.StatsTimeout)
	}
	if cfg.SpreadChecks != 5 {
		t.Fatalf("expected SpreadChecks 5, got %d", cfg.SpreadChecks)
	}

	if cfg.isEmpty() {
		t.Fatalf("expected populated GlobalConfig not to be reported as empty")
	}
}

func TestMapDefaultsFromAPIAndIsEmpty(t *testing.T) {
	// Empty map should produce an "empty" defaults config.
	emptyCfg := mapDefaultsFromAPI(map[string]interface{}{})
	if !emptyCfg.isEmpty() {
		t.Fatalf("expected empty DefaultsConfig to be reported as empty")
	}

	input := map[string]interface{}{
		"name":            "unnamed_defaults_1",
		"mode":            "http",
		"timeout_client":  "30s",
		"timeout_server":  "30s",
		"timeout_connect": "5s",
		"timeout_queue":   "5s",
		"timeout_tunnel":  "60s",
		"balance":         "roundrobin",
		"log":             "global",
	}

	cfg := mapDefaultsFromAPI(input)

	if cfg.APIVersion != "haproxyctl/v1" {
		t.Fatalf("expected APIVersion haproxyctl/v1, got %s", cfg.APIVersion)
	}
	if cfg.Kind != kindDefaults {
		t.Fatalf("expected Kind %s, got %s", kindDefaults, cfg.Kind)
	}
	if cfg.Name != "unnamed_defaults_1" {
		t.Fatalf("unexpected Name: %s", cfg.Name)
	}
	if cfg.Mode != "http" {
		t.Fatalf("unexpected Mode: %s", cfg.Mode)
	}
	if cfg.TimeoutClient != "30s" ||
		cfg.TimeoutServer != "30s" ||
		cfg.TimeoutConnect != "5s" ||
		cfg.TimeoutQueue != "5s" ||
		cfg.TimeoutTunnel != "60s" {
		t.Fatalf("unexpected timeout values: %+v", cfg)
	}
	if cfg.Balance != "roundrobin" {
		t.Fatalf("unexpected Balance: %s", cfg.Balance)
	}
	if cfg.Log != "global" {
		t.Fatalf("unexpected Log: %s", cfg.Log)
	}

	if cfg.isEmpty() {
		t.Fatalf("expected populated DefaultsConfig not to be reported as empty")
	}
}
