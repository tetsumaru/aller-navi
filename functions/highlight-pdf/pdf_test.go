package highlightpdf

import (
	"strings"
	"testing"
)

func TestPolyBoundsNil(t *testing.T) {
	x1, y1, x2, y2 := polyBounds(nil, 100, 200)
	if x1 != 0 || y1 != 0 || x2 != 0 || y2 != 0 {
		t.Errorf("expected all zeros for nil poly, got (%.1f,%.1f,%.1f,%.1f)", x1, y1, x2, y2)
	}
}

func TestContainsTarget(t *testing.T) {
	tests := []struct {
		name   string
		text   string
		target string
		want   bool
	}{
		{"contains target", "卵焼き", "卵", true},
		{"does not contain target", "ご飯", "卵", false},
		{"empty text", "", "卵", false},
		{"empty target", "卵焼き", "", true}, // strings.Contains("卵焼き", "") == true
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := strings.Contains(tc.text, tc.target)
			if got != tc.want {
				t.Errorf("strings.Contains(%q, %q) = %v, want %v", tc.text, tc.target, got, tc.want)
			}
		})
	}
}
