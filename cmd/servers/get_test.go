package servers

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"haproxyctl/internal"

	"github.com/spf13/cobra"
)

func TestGetServers_BackendNotFound(t *testing.T) {
	origGet := getServersBackendResource
	defer func() { getServersBackendResource = origGet }()

	getServersBackendResource = func(_ string) (map[string]interface{}, error) {
		return nil, errors.New("HAProxy API error (404): missing object")
	}

	output := internal.CaptureStderr(t, func() {
		cmd := &cobra.Command{}
		cmd.SetArgs([]string{})
		getServers(cmd, "missing-backend", "")
	})
	want := `Error: backend "missing-backend" not found`
	if !bytes.Contains([]byte(output), []byte(want)) {
		t.Fatalf("expected stderr to contain %q, got: %s", want, output)
	}
}

func TestGetServers_ServerNotFound(t *testing.T) {
	origGet := getServersBackendResource
	origSend := sendServersRequest
	defer func() {
		getServersBackendResource = origGet
		sendServersRequest = origSend
	}()

	getServersBackendResource = func(_ string) (map[string]interface{}, error) {
		return map[string]interface{}{}, nil
	}
	sendServersRequest = func(_ context.Context, _, _ string, _ map[string]string, _ interface{}) ([]byte, error) {
		return nil, errors.New("HAProxy API error (404): missing object")
	}

	output := internal.CaptureStdout(t, func() {
		cmd := &cobra.Command{}
		getServers(cmd, "backend1", "missing-server")
	})
	want := "server/backend1/missing-server not found"
	if !bytes.Contains([]byte(output), []byte(want)) {
		t.Fatalf("expected stdout to contain %q, got: %s", want, output)
	}
}

func TestGetServers_ArgsValidation(t *testing.T) {
	output := internal.CaptureStderr(t, func() {
		cmd := GetServersCmd
		cmd.SetArgs([]string{})
		_ = cmd.Execute()
	})

	if !strings.Contains(output, "accepts between 1 and 2 arg(s), received 0") {
		t.Fatalf("expected error about arg count, got:\n%s", output)
	}
	if !strings.Contains(output, "servers <backend_name> [server_name] [flags]") {
		t.Fatalf("expected usage for servers command, got:\n%s", output)
	}
}
