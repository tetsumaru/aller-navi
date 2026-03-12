package highlightpdf

import (
	"testing"
)

func TestContainsAny(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		allergens []string
		want      bool
	}{
		{
			name:      "contains allergen",
			text:      "卵焼き",
			allergens: []string{"卵", "乳"},
			want:      true,
		},
		{
			name:      "contains one of multiple allergens",
			text:      "牛乳スープ",
			allergens: []string{"卵", "乳"},
			want:      true,
		},
		{
			name:      "contains no allergen",
			text:      "ご飯",
			allergens: []string{"卵", "乳", "小麦"},
			want:      false,
		},
		{
			name:      "empty allergens",
			text:      "卵焼き",
			allergens: []string{},
			want:      false,
		},
		{
			name:      "empty text",
			text:      "",
			allergens: []string{"卵"},
			want:      false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := containsAny(tc.text, tc.allergens)
			if got != tc.want {
				t.Errorf("containsAny(%q, %v) = %v, want %v", tc.text, tc.allergens, got, tc.want)
			}
		})
	}
}

func TestPolyBoundsNil(t *testing.T) {
	x1, y1, x2, y2 := polyBounds(nil)
	if x1 != 0 || y1 != 0 || x2 != 0 || y2 != 0 {
		t.Errorf("expected all zeros for nil poly, got (%.1f,%.1f,%.1f,%.1f)", x1, y1, x2, y2)
	}
}
