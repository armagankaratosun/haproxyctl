package servers

import (
	"strings"
	"testing"

	"haproxyctl/internal"
)

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
