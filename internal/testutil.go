package internal

import (
	"bytes"
	"os"
	"testing"
)

const testLogOutputEnv = "HAPROXYCTL_TEST_LOG_OUTPUT"

// CaptureStdout runs fn while capturing everything written to os.Stdout.
// It returns the captured output as a string. When the HAPROXYCTL_TEST_LOG_OUTPUT
// environment variable is set, the captured output is also logged via t.Logf.
func CaptureStdout(t *testing.T, fn func()) string {
	t.Helper()

	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdout = w

	fn()

	if err := w.Close(); err != nil {
		t.Fatalf("failed to close writer: %v", err)
	}
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("failed to read from pipe: %v", err)
	}
	os.Stdout = origStdout

	out := buf.String()
	if os.Getenv(testLogOutputEnv) != "" {
		t.Logf("captured stdout:\n%s", out)
	}
	return out
}

// CaptureStderr runs fn while capturing everything written to os.Stderr.
// It returns the captured output as a string. When the HAPROXYCTL_TEST_LOG_OUTPUT
// environment variable is set, the captured output is also logged via t.Logf.
func CaptureStderr(t *testing.T, fn func()) string {
	t.Helper()

	origStderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stderr = w

	fn()

	if err := w.Close(); err != nil {
		t.Fatalf("failed to close writer: %v", err)
	}
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("failed to read from pipe: %v", err)
	}
	os.Stderr = origStderr

	out := buf.String()
	if os.Getenv(testLogOutputEnv) != "" {
		t.Logf("captured stderr:\n%s", out)
	}
	return out
}
