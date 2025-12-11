package internal

import "testing"

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
