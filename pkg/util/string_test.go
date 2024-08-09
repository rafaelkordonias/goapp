package util

import (
	"testing"
)

func TestRandString(t *testing.T) {
	length := 10

	// Check that length of result matches the requested
	result := RandString(length)
	if len(result) != length {
		t.Errorf("Expected length %d, got %d", length, len(result))
	}

	// Check that contains only hex chars
	for _, char := range result {
		if (char < '0' || char > '9') && (char < 'A' || char > 'F') {
			t.Errorf("Unexpected character %c in result", char)
		}
	}
}

func BenchmarkRandString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RandString(10)
	}
}
