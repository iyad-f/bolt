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

// Header represents HTTP header key-value pairs.
// Keys are canonicalized to Title-Case (e.g., "Content-Type").
type Header map[string][]string

// Get returns the first value for the given key, or "" if not present.
func (h Header) Get(key string) string {
	key = CanonicalHeaderKey(key)
	if values, ok := h[key]; ok && len(values) > 0 {
		return values[0]
	}

	return ""
}

// Set sets the key to a single value, replacing any existing values.
func (h Header) Set(key, value string) {
	h[CanonicalHeaderKey(key)] = []string{value}
}

// Add appends a value to the given key.
func (h Header) Add(key, value string) {
	key = CanonicalHeaderKey(key)
	h[key] = append(h[key], value)
}

// Del removes the given key and its values.
func (h Header) Del(key string) {
	delete(h, CanonicalHeaderKey(key))
}

// Clone returns a deep copy of the header.
func (h Header) Clone() Header {
	clone := Header{}

	for key, values := range h {
		newValues := make([]string, len(values))
		copy(newValues, values)
		clone[key] = newValues
	}

	return clone
}

func isUpper(b byte) bool { return b >= 'A' && b <= 'Z' }

func isLower(b byte) bool { return b >= 'a' && b <= 'z' }

func toUpper(b byte) byte {
	if isLower(b) {
		b -= 32
	}
	return b
}

func toLower(b byte) byte {
	if isUpper(b) {
		b += 32
	}
	return b
}

// isTokenByte reports whether b is a valid HTTP header field name byte (RFC 7230 token).
func isTokenByte(b byte) bool {
	if isLower(b) || isUpper(b) || b >= '0' && b <= '9' {
		return true
	}
	switch b {
	case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~':
		return true
	}
	return false
}

// CanonicalHeaderKey returns the canonical format of the header key s.
// It converts the first letter and any letter following a hyphen to uppercase,
// and the rest to lowercase. For example, "content-type" becomes "Content-Type".
// If s contains invalid header key bytes, it is returned unchanged.
func CanonicalHeaderKey(s string) string {
	// Check if the string is already canonical or has a invalid token.
	canonical := true
	upper := true
	for i := range s {
		currByte := s[i]
		if !isTokenByte(currByte) && currByte != '-' {
			return s
		}
		if canonical {
			if upper && isLower(currByte) {
				canonical = false
			}
			if !upper && isUpper(currByte) {
				canonical = false
			}
		}
		upper = currByte == '-'
	}
	if canonical {
		return s
	}

	// At this point its not canonical and has valid tokens then do a allocation and make it canonical.
	byteArray := []byte(s)
	upper = true
	for i := range byteArray {
		if upper {
			byteArray[i] = toUpper(byteArray[i])
		} else {
			byteArray[i] = toLower(byteArray[i])
		}
		upper = byteArray[i] == '-'
	}

	return string(byteArray)
}
