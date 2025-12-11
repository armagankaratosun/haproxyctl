package internal

import "testing"

// TestCompatLiveDataPlaneAPI performs an optional integration-style smoke test
// against the configured HAProxy Data Plane API. It is intentionally skipped
// when no local haproxyctl config is available, so it does not affect normal
// unit-test runs on machines without a running API.
func TestCompatLiveDataPlaneAPI(t *testing.T) {
	t.Parallel()

	// If we cannot load configuration, assume this environment does not have
	// a running Data Plane API and skip the test instead of failing.
	if _, err := LoadConfig(); err != nil {
		t.Skipf("skipping compat test; failed to load haproxyctl config: %v", err)
	}

	version, err := GetConfigurationVersion()
	if err != nil {
		t.Fatalf("GetConfigurationVersion failed: %v", err)
	}
	if version < 0 {
		t.Fatalf("unexpected negative configuration version: %d", version)
	}

	// Basic list endpoints should be reachable and parseable.
	if _, err := GetResourceList("/services/haproxy/configuration/backends"); err != nil {
		t.Fatalf("failed to list backends: %v", err)
	}

	if _, err := GetResourceList("/services/haproxy/configuration/frontends"); err != nil {
		t.Fatalf("failed to list frontends: %v", err)
	}
}
