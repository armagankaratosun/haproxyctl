package internal

import "testing"

func TestNormalizeAPIBaseURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "no version",
			input: "http://example:5555",
			want:  "http://example:5555/v3",
		},
		{
			name:  "trailing slash",
			input: "http://example:5555/",
			want:  "http://example:5555/v3",
		},
		{
			name:  "explicit v1",
			input: "http://example:5555/v1",
			want:  "http://example:5555/v1",
		},
		{
			name:  "explicit v3 with extra slash and spaces",
			input: "  http://example:5555/v3/  ",
			want:  "http://example:5555/v3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := normalizeAPIBaseURL(tt.input)
			if got != tt.want {
				t.Fatalf("normalizeAPIBaseURL(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
