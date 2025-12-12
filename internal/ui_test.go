package internal

import (
	"bytes"
	"errors"
	"testing"
)

func TestResourceID(t *testing.T) {
	tests := []struct {
		kind string
		name string
		want string
	}{
		{"Backend", "mybackend", "backend/mybackend"},
		{"Server", "backend1/server1", "server/backend1/server1"},
		{"Defaults", "config", "defaults/config"},
	}

	for _, tt := range tests {
		got := ResourceID(tt.kind, tt.name)
		if got != tt.want {
			t.Fatalf("ResourceID(%q, %q) = %q, want %q", tt.kind, tt.name, got, tt.want)
		}
	}
}

func TestPrintStatus(t *testing.T) {
	output := CaptureStdout(t, func() {
		PrintStatus("Backend", "mybackend", ActionCreated)
	})
	want := "backend/mybackend created"
	if !bytes.Contains([]byte(output), []byte(want)) {
		t.Fatalf("expected output to contain %q, got: %s", want, output)
	}
}

func TestPrintDryRun(t *testing.T) {
	output := CaptureStdout(t, func() {
		PrintDryRun()
	})
	want := "Dry run mode enabled. No changes made."
	if !bytes.Contains([]byte(output), []byte(want)) {
		t.Fatalf("expected output to contain %q, got: %s", want, output)
	}
}

func TestFormatAPIErrorAndWrap(t *testing.T) {
	baseErr := errors.New("HAProxy API error (404): missing object")

	err := FormatAPIError("Backend", "foo", "delete", baseErr)
	if err == nil || err.Error() != "backend \"foo\" not found" {
		t.Fatalf("unexpected FormatAPIError result: %v", err)
	}

	wrapped := WrapIfAPIError("Backend", "foo", "delete", nil)
	if wrapped != nil {
		t.Fatalf("expected WrapIfAPIError to return nil when err is nil, got: %v", wrapped)
	}
}
