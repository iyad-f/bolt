package bolt

import "testing"

func TestCanonicalHeaderKey(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"content-type", "Content-Type"},
		{"Content-Type", "Content-Type"},
		{"CONTENT-TYPE", "Content-Type"},
		{"cOnTeNt-TyPe", "Content-Type"},
		{"x-forwarded-for", "X-Forwarded-For"},
		{"host", "Host"},
		{"x", "X"},
		{"", ""},
		{"invalid header", "invalid header"},
	}
	for _, tt := range tests {
		got := CanonicalHeaderKey(tt.input)
		if got != tt.want {
			t.Errorf("CanonicalHeaderKey(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
