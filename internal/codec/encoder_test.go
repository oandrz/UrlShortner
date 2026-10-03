package codec

import (
	"math"
	"strings"
	"testing"
)

func TestEncodeBase62(t *testing.T) {
	tests := []struct {
		input uint64
		want  string
	}{
		{0, "0"},
		{1, "1"},
		{61, "Z"},
		{62, "10"},
		{1000, "g8"},
		{3843, "ZZ"},
		{3844, "100"},
		{1000000, "4c92"},
	}

	for _, tt := range tests {
		got := EncodeBase62(tt.input)
		if got != tt.want {
			t.Errorf("EncodeBase62(%d) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestDecodeBase62(t *testing.T) {
	tests := []struct {
		input string
		want  uint64
	}{
		{"0", 0},
		{"1", 1},
		{"g", 16},
		{"Z", 61},
		{"10", 62},
		{"11", 63},
		{"g8", 1000},
		{"100", 3844},
		{"4c92", 1000000},
	}

	for _, tt := range tests {
		got := decodeBase62(tt.input)
		if got != tt.want {
			t.Errorf("decodeBase62(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestBase62RoundTrip(t *testing.T) {
	inputs := []uint64{0, 1, 61, 62, 3843, 3844, 1000000, math.MaxUint64}

	for _, n := range inputs {
		code := EncodeBase62(n)
		got := decodeBase62(code)
		if got != n {
			t.Errorf("decodeBase62(EncodeBase62(%d)) = %d (code %q), want %d", n, got, code, n)
		}
	}
}

func TestEncodeBase62Alphabet(t *testing.T) {
	inputs := []uint64{math.MaxUint64}
	for n := uint64(0); n <= 10000; n++ {
		inputs = append(inputs, n)
	}

	for _, n := range inputs {
		code := EncodeBase62(n)
		for _, r := range code {
			if !strings.ContainsRune(alphabet, r) {
				t.Errorf("EncodeBase62(%d) = %q, contains %q which is not in the alphabet", n, code, r)
			}
		}
	}
}

func TestEncodeBase62Length(t *testing.T) {
	// 62^k is the smallest number needing k+1 chars; 62^k - 1 is the largest fitting in k.
	n := uint64(1)
	for k := 0; k <= 10; k++ {
		if got := len(EncodeBase62(n)); got != k+1 {
			t.Errorf("len(EncodeBase62(62^%d)) = %d, want %d", k, got, k+1)
		}
		if k > 0 {
			if got := len(EncodeBase62(n - 1)); got != k {
				t.Errorf("len(EncodeBase62(62^%d - 1)) = %d, want %d", k, got, k)
			}
		}
		n *= 62
	}
}
