package backends

import "testing"

func TestBackendWithServersToPayload_TimeoutsAndFlags(t *testing.T) {
	t.Parallel()

	b := &backendWithServers{
		backendConfig: backendConfig{
			Name:                 "test-backend",
			TimeoutClient:        "30s",
			TimeoutHTTPKeepAlive: "10s",
			TimeoutHTTPRequest:   "1500ms",
			TimeoutQueue:         "",
			TimeoutServer:        "5000", // already milliseconds
			TimeoutServerFin:     "2s",
			TCPKA:                true,
			Redispatch:           true,
		},
	}

	payload := b.toPayload()

	if payload.Name != "test-backend" {
		t.Fatalf("payload.Name = %q, want %q", payload.Name, "test-backend")
	}

	if payload.TimeoutClient != 30000 {
		t.Fatalf("TimeoutClient = %d, want %d", payload.TimeoutClient, 30000)
	}
	if payload.TimeoutHTTPKeepAlive != 10000 {
		t.Fatalf("TimeoutHTTPKeepAlive = %d, want %d", payload.TimeoutHTTPKeepAlive, 10000)
	}
	if payload.TimeoutHTTPRequest != 1500 {
		t.Fatalf("TimeoutHTTPRequest = %d, want %d", payload.TimeoutHTTPRequest, 1500)
	}
	if payload.TimeoutQueue != 0 {
		t.Fatalf("TimeoutQueue = %d, want %d", payload.TimeoutQueue, 0)
	}
	if payload.TimeoutServer != 5000 {
		t.Fatalf("TimeoutServer = %d, want %d", payload.TimeoutServer, 5000)
	}
	if payload.TimeoutServerFin != 2000 {
		t.Fatalf("TimeoutServerFin = %d, want %d", payload.TimeoutServerFin, 2000)
	}

	if payload.TCPKA != stateEnabled {
		t.Fatalf("TCPKA = %q, want %q", payload.TCPKA, stateEnabled)
	}
	if payload.Redispatch == nil {
		t.Fatal("Redispatch is nil, want non-nil")
	}
	if payload.Redispatch.Enabled != stateEnabled {
		t.Fatalf("Redispatch.Enabled = %q, want %q", payload.Redispatch.Enabled, stateEnabled)
	}
}
