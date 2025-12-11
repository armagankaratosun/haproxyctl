// Package internal contains shared helpers for haproxyctl.
package internal

import (
	"fmt"
	"log"
	"os"
	"strings"
)

const (
	// ActionCreated indicates a resource was created.
	ActionCreated = "created"
	// ActionConfigured indicates a resource was updated or configured.
	ActionConfigured = "configured"
	// ActionUnchanged indicates a resource was already up to date.
	ActionUnchanged = "unchanged"
	// ActionDeleted indicates a resource was deleted.
	ActionDeleted = "deleted"
)

// ResourceID builds a kubectl-like identifier such as "backend/example-backend".
func ResourceID(kind, name string) string {
	return fmt.Sprintf("%s/%s", strings.ToLower(kind), name)
}

// PrintStatus prints a concise status line for a resource, for example:
// "backend/example-backend created".
func PrintStatus(kind, name, action string) {
	if _, err := fmt.Fprintf(os.Stdout, "%s %s\n", ResourceID(kind, name), action); err != nil {
		log.Printf("warning: failed to write status for %s %q: %v", strings.ToLower(kind), name, err)
	}
}

// PrintDryRun prints a standard dry‑run message.
func PrintDryRun() {
	if _, err := fmt.Fprintln(os.Stdout, "Dry run mode enabled. No changes made."); err != nil {
		log.Printf("warning: failed to write dry-run message: %v", err)
	}
}

// IsAlreadyExistsError reports whether the given error corresponds to
// a 409 Already Exists response from the Data Plane API.
func IsAlreadyExistsError(err error) bool {
	if err == nil {
		return false
	}
	// SendRequest formats errors as: "HAProxy API error (%d): ..."
	return strings.Contains(err.Error(), "HAProxy API error (409)")
}

// FormatAPIError normalizes HAProxy API errors into user‑friendly messages.
// It recognises common cases like 404 and 409 and falls back to a generic
// description otherwise.
func FormatAPIError(kind, name, operation string, err error) error {
	if err == nil {
		return nil
	}

	lowerKind := strings.ToLower(kind)

	if IsAlreadyExistsError(err) {
		return fmt.Errorf("%s %q already exists (consider using 'haproxyctl apply -f ...')", lowerKind, name)
	}
	if IsNotFoundError(err) {
		return fmt.Errorf("%s %q not found", lowerKind, name)
	}

	// Preserve original context for unexpected errors.
	return fmt.Errorf("HAProxy API %s %s: %w", operation, ResourceID(kind, name), err)
}

// WrapIfAPIError applies FormatAPIError only when err is non‑nil.
func WrapIfAPIError(kind, name, operation string, err error) error {
	if err == nil {
		return nil
	}
	return FormatAPIError(kind, name, operation, err)
}
