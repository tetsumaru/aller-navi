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

// TestMatchesAnyTargetMultiWord は部分一致マッチングの動作を確認する。
func TestMatchesAnyTargetMultiWord(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		targets []string
		want    bool
	}{
		{
			name:    "部分一致でマッチする",
			text:    "主食バターロール",
			targets: []string{"主食バターロール"},
			want:    true,
		},
		{
			name:    "単語単独ではマッチしない（後半）",
			text:    "バターロール",
			targets: []string{"主食バターロール"},
			want:    false,
		},
		{
			name:    "部分一致する単一単語ターゲット",
			text:    "バターロール",
			targets: []string{"バターロール"},
			want:    true,
		},
		{
			name:    "牛乳を含む単語もマッチする（既存の単語レベル動作）",
			text:    "牛乳未満児100cc",
			targets: []string{"牛乳"},
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

// TestExactMatchesAnyTarget は複数単語スパンの完全一致マッチングを確認する。
// 「牛乳」のような単一単語ターゲットが隣接単語との結合でマッチしないことを保証する。
func TestExactMatchesAnyTarget(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		targets []string
		want    bool
	}{
		{
			name:    "完全一致でマッチする",
			text:    "主食バターロール",
			targets: []string{"主食バターロール"},
			want:    true,
		},
		{
			name:    "牛乳を含む複数単語スパンはマッチしない",
			text:    "鶏こし豆腐牛乳",
			targets: []string{"牛乳"},
			want:    false,
		},
		{
			name:    "牛乳未満児100ccはマッチしない",
			text:    "牛乳未満児100cc",
			targets: []string{"牛乳"},
			want:    false,
		},
		{
			name:    "単一単語の完全一致",
			text:    "牛乳",
			targets: []string{"牛乳"},
			want:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := exactMatchesAnyTarget(tc.text, tc.targets)
			if got != tc.want {
				t.Errorf("exactMatchesAnyTarget(%q, %v) = %v, want %v", tc.text, tc.targets, got, tc.want)
			}
		})
	}
}
