package frontends

import "testing"

func TestFrontendWithBindsToPayload_Timeouts(t *testing.T) {
	t.Parallel()

	f := &frontendWithBinds{
		frontendConfig: frontendConfig{
			Name:                 "test-frontend",
			TimeoutClient:        "30s",
			TimeoutHTTPRequest:   "1500ms",
			TimeoutHTTPKeepAlive: "",
			TimeoutQueue:         "1s",
			TimeoutServer:        "5000",
		},
	}

	payload := f.ToPayload()

	if payload.Name != "test-frontend" {
		t.Fatalf("payload.Name = %q, want %q", payload.Name, "test-frontend")
	}

	if payload.TimeoutClient != 30000 {
		t.Fatalf("TimeoutClient = %d, want %d", payload.TimeoutClient, 30000)
	}
	if payload.TimeoutHTTPRequest != 1500 {
		t.Fatalf("TimeoutHTTPRequest = %d, want %d", payload.TimeoutHTTPRequest, 1500)
	}
	if payload.TimeoutHTTPKeepAlive != 0 {
		t.Fatalf("TimeoutHTTPKeepAlive = %d, want %d", payload.TimeoutHTTPKeepAlive, 0)
	}
	if payload.TimeoutQueue != 1000 {
		t.Fatalf("TimeoutQueue = %d, want %d", payload.TimeoutQueue, 1000)
	}
	if payload.TimeoutServer != 5000 {
		t.Fatalf("TimeoutServer = %d, want %d", payload.TimeoutServer, 5000)
	}
}
