// Copyright 2026 Iyad
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

func TestHeaderGet(t *testing.T) {
	h := Header{
		"Content-Type": []string{"application/json"},
		"Set-Cookie":   []string{"a=1", "b=2"},
	}

	tests := []struct {
		input string
		want  string
	}{
		{"Content-Type", "application/json"},
		{"content-type", "application/json"},
		{"Set-Cookie", "a=1"},
		{"Missing-Key", ""},
	}
	for _, tt := range tests {
		got := h.Get(tt.input)
		if got != tt.want {
			t.Errorf("Get(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestHeaderSet(t *testing.T) {
	h := Header{}

	h.Set("Content-Type", "text/html")
	if got := h.Get("Content-Type"); got != "text/html" {
		t.Errorf("after Set, Get = %q, want %q", got, "text/html")
	}

	h.Set("Content-Type", "application/json")
	if got := h.Get("Content-Type"); got != "application/json" {
		t.Errorf("after overwrite, Get = %q, want %q", got, "application/json")
	}
}

func TestHeaderAdd(t *testing.T) {
	h := Header{}

	h.Add("Set-Cookie", "a=1")
	h.Add("Set-Cookie", "b=2")

	values := h["Set-Cookie"]
	if len(values) != 2 {
		t.Fatalf("expected 2 values, got %d", len(values))
	}
	if values[0] != "a=1" || values[1] != "b=2" {
		t.Errorf("values = %v, want [a=1 b=2]", values)
	}
}

func TestHeaderDel(t *testing.T) {
	h := Header{}

	h.Set("Content-Type", "text/html")
	h.Del("Content-Type")

	if got := h.Get("Content-Type"); got != "" {
		t.Errorf("after Del, Get = %q, want empty", got)
	}
}

func TestHeaderClone(t *testing.T) {
	h := Header{}
	h.Set("Content-Type", "text/html")
	h.Add("Set-Cookie", "a=1")

	clone := h.Clone()

	clone.Set("Content-Type", "application/json")
	clone.Add("Set-Cookie", "b=2")

	if got := h.Get("Content-Type"); got != "text/html" {
		t.Errorf("original changed after clone modify: got %q", got)
	}
	if len(h["Set-Cookie"]) != 1 {
		t.Errorf("original Set-Cookie has %d values, want 1", len(h["Set-Cookie"]))
	}
}
