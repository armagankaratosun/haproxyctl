package frontends

import (
	"bytes"
	"errors"
	"testing"

	"haproxyctl/internal"

	"github.com/spf13/cobra"
)

func TestGetFrontends_NotFound(t *testing.T) {
	origGet := getFrontendsResource
	defer func() { getFrontendsResource = origGet }()

	getFrontendsResource = func(_ string) (map[string]interface{}, error) {
		return nil, errors.New("HAProxy API error (404): missing object")
	}

	output := internal.CaptureStdout(t, func() {
		cmd := &cobra.Command{}
		getFrontends(cmd, "missing-frontend")
	})

	want := "frontend/missing-frontend not found"
	if !bytes.Contains([]byte(output), []byte(want)) {
		t.Fatalf("expected output to contain %q, got: %s", want, output)
	}
}
