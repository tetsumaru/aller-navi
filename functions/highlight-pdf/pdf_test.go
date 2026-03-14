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

// TestMatchesAnyTargetMultiWord は、Vision API が複数単語に分割したテキストでも
// 結合済みパラグラフブロックによってマッチすることを確認する。
// 例: 「主食バターロール」が「主食」「バターロール」に分割されるケース。
func TestMatchesAnyTargetMultiWord(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		targets []string
		want    bool
	}{
		{
			name:    "単語が結合されてマッチする",
			text:    "主食バターロール",
			targets: []string{"主食バターロール"},
			want:    true,
		},
		{
			name:    "単語単独ではマッチしない",
			text:    "バターロール",
			targets: []string{"主食バターロール"},
			want:    false,
		},
		{
			name:    "単語単独ではマッチしない（前半）",
			text:    "主食",
			targets: []string{"主食バターロール"},
			want:    false,
		},
		{
			name:    "部分一致する単一単語ターゲット",
			text:    "バターロール",
			targets: []string{"バターロール"},
			want:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := matchesAnyTarget(tc.text, tc.targets)
			if got != tc.want {
				t.Errorf("matchesAnyTarget(%q, %v) = %v, want %v", tc.text, tc.targets, got, tc.want)
			}
		})
	}
}
