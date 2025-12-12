package acls

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"haproxyctl/internal"

	"github.com/spf13/cobra"
)

func TestGetACLs_FrontendNotFound(t *testing.T) {
	orig := getACLsRequest
	defer func() { getACLsRequest = orig }()

	getACLsRequest = func(_ context.Context, _, _ string, _ map[string]string, _ interface{}) ([]byte, error) {
		return nil, errors.New("HAProxy API error (404): missing object")
	}

	output := internal.CaptureStderr(t, func() {
		cmd := &cobra.Command{}
		getACLs("missing-frontend", cmd)
	})
	want := `Error: frontend "missing-frontend" not found`
	if !bytes.Contains([]byte(output), []byte(want)) {
		t.Fatalf("expected stderr to contain %q, got: %s", want, output)
	}
}

func TestGetACLs_ArgsValidation(t *testing.T) {
	output := internal.CaptureStderr(t, func() {
		cmd := GetACLsCmd
		cmd.SetArgs([]string{})
		_ = cmd.Execute()
	})

	if !strings.Contains(output, "accepts 1 arg(s), received 0") {
		t.Fatalf("expected error about arg count, got:\n%s", output)
	}
	if !strings.Contains(output, "acls <frontend_name>") {
		t.Fatalf("expected usage for acls command, got:\n%s", output)
	}
}
