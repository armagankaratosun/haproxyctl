package backends

import (
	"bytes"
	"errors"
	"testing"

	"haproxyctl/internal"

	"github.com/spf13/cobra"
)

// TestGetBackends_NotFound verifies that a 404 when fetching a single backend
// results in a friendly "<kind>/<name> not found" message instead of a raw
// API error dump.
func TestGetBackends_NotFound(t *testing.T) {
	origGet := getBackendsResource
	defer func() { getBackendsResource = origGet }()

	getBackendsResource = func(_ string) (map[string]interface{}, error) {
		return nil, errors.New("HAProxy API error (404): missing object")
	}

	output := internal.CaptureStdout(t, func() {
		cmd := &cobra.Command{}
		getBackends(cmd, "missing-backend")
	})

	want := "backend/missing-backend not found"
	if !bytes.Contains([]byte(output), []byte(want)) {
		t.Fatalf("expected output to contain %q, got: %s", want, output)
	}
}
