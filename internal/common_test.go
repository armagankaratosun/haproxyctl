package internal

import (
	"strings"
	"testing"
)

func TestParseDurationToMillis(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		{name: "empty", input: "", want: 0, wantErr: false},
		{name: "plain milliseconds", input: "1500", want: 1500, wantErr: false},
		{name: "seconds", input: "30s", want: 30000, wantErr: false},
		{name: "milliseconds suffix", input: "500ms", want: 500, wantErr: false},
		{name: "trimmed spaces", input: " 15s ", want: 15000, wantErr: false},
		{name: "invalid duration", input: "not-a-duration", want: 0, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ParseDurationToMillis(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParseDurationToMillis(%q) expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseDurationToMillis(%q) unexpected error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Fatalf("ParseDurationToMillis(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestFormatMillisAsDuration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input int
		want  string
	}{
		{name: "zero", input: 0, want: ""},
		{name: "seconds", input: 30000, want: "30s"},
		{name: "fractional seconds", input: 1500, want: "1.5s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := FormatMillisAsDuration(tt.input)
			if got != tt.want {
				t.Fatalf("FormatMillisAsDuration(%d) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestSortByStringField(t *testing.T) {
	t.Parallel()

	list := []map[string]interface{}{
		{"name": "b", "other": 1},
		{"name": "a", "other": 2},
		{"other": 3}, // no name field
	}

	SortByStringField(list, "name")

	// Entries without the field are treated as empty string and appear first.
	if _, ok := list[0]["name"]; ok {
		t.Fatalf("expected first element to have no name field, got %+v", list[0])
	}
	if list[1]["name"] != "a" {
		t.Fatalf("expected second element to have name 'a', got %+v", list[1])
	}
	if list[2]["name"] != "b" {
		t.Fatalf("expected third element to have name 'b', got %+v", list[2])
	}
}

func TestGetSortedKeys(t *testing.T) {
	t.Parallel()

	row := map[string]interface{}{
		"mode": "http",
		"name": "backend1",
		"log":  "global",
	}

	keys := getSortedKeys(row)

	if len(keys) != 3 {
		t.Fatalf("expected 3 keys, got %d", len(keys))
	}
	if keys[0] != "name" {
		t.Fatalf("expected primary key 'name', got %s", keys[0])
	}
	if !(keys[1] == "log" && keys[2] == "mode" || keys[1] == "mode" && keys[2] == "log") {
		t.Fatalf("unexpected secondary key order: %v", keys[1:])
	}
}

func TestFormatOutput_TableAndYAML(t *testing.T) {
	// First, table output for a single object.
	row := map[string]interface{}{
		"name": "backend1",
		"mode": "http",
	}

	output := CaptureStdout(t, func() {
		FormatOutput(row, "")

		// Then, YAML output for a manifest list.
		list := ManifestList{
			APIVersion: "haproxyctl/v1",
			Kind:       "List",
			Items:      []interface{}{row},
		}
		FormatOutput(list, OutputFormatYAML)
	})

	if !strings.Contains(output, "NAME") || !strings.Contains(output, "backend1") {
		t.Fatalf("expected table output with NAME and backend1, got:\n%s", output)
	}
	if !strings.Contains(output, "apiVersion: haproxyctl/v1") || !strings.Contains(output, "kind: List") {
		t.Fatalf("expected YAML manifest list output, got:\n%s", output)
	}
}
