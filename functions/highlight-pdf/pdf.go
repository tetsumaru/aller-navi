package highlightpdf

import (
	"bytes"
	"fmt"
	"log/slog"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/color"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

// highlightColor はハイライトに適用する色（黄色）です。
var highlightColor = color.SimpleColor{R: 1.0, G: 0.95, B: 0.0}

// ProcessPDF は PDF に target 文字列を含むテキストブロックのハイライトを追加します。
// pages は PDF のページと 1:1 対応している必要があります（インデックス 0 = 1 ページ目）。
func ProcessPDF(pdfBytes []byte, pages []PageInfo, target string) ([]byte, error) {
	// PDF の各ページサイズを取得する。
	dims, err := api.PageDims(bytes.NewReader(pdfBytes), nil)
	if err != nil {
		return nil, fmt.Errorf("get page dims: %w", err)
	}

	// ページ番号 → アノテーション のマップを構築して一括追加に備える。
	annotationsMap := make(map[int][]model.AnnotationRenderer)

	for pageIdx, page := range pages {
		if pageIdx >= len(dims) {
			break
		}
		pdfW := dims[pageIdx].Width
		pdfH := dims[pageIdx].Height

		if page.Width == 0 || page.Height == 0 {
			continue
		}
		scaleX := pdfW / float64(page.Width)
		scaleY := pdfH / float64(page.Height)

		pageNum := pageIdx + 1 // pdfcpu はページ番号が 1 始まり

		targets := splitTargets(target)
		for _, block := range page.Blocks {
			if !matchesAnyTarget(block.Text, targets) {
				continue
			}

			// 画像ピクセル座標（左上原点）を PDF ポイント座標（左下原点）に変換する。
			pdfX1 := block.X1 * scaleX
			pdfY1 := pdfH - (block.Y2 * scaleY) // Y 軸を反転
			pdfX2 := block.X2 * scaleX
			pdfY2 := pdfH - (block.Y1 * scaleY) // Y 軸を反転

			rect := types.NewRectangle(pdfX1, pdfY1, pdfX2, pdfY2)

			ann := model.NewSquareAnnotation(
				*rect,
				"",                   // contents
				"",                   // id
				0,                    // AnnotationFlags
				0,                    // borderWidth (no border)
				model.BSSolid,        // borderStyle
				nil,                  // borderCol (no border)
				false,                // cloudyBorder
				0,                    // cloudyBorderIntensity
				&highlightColor,      // fillCol
				0, 0, 0, 0,           // MLeft, MTop, MRight, MBot
			)

			slog.Info("adding highlight",
				"page", pageNum,
				"text", block.Text,
				"rect", fmt.Sprintf("(%.1f,%.1f)-(%.1f,%.1f)", pdfX1, pdfY1, pdfX2, pdfY2),
			)

			annotationsMap[pageNum] = append(annotationsMap[pageNum], ann)
		}
	}

	if len(annotationsMap) == 0 {
		return pdfBytes, nil
	}

	var buf bytes.Buffer
	if err := api.AddAnnotationsMap(bytes.NewReader(pdfBytes), &buf, annotationsMap, nil); err != nil {
		return nil, fmt.Errorf("add annotations: %w", err)
	}

	return buf.Bytes(), nil
}

// splitTargets は改行区切りの target 文字列を個別のアレルゲン文字列のスライスに変換します。
func splitTargets(target string) []string {
	var result []string
	for _, t := range strings.Split(target, "\n") {
		if t = strings.TrimSpace(t); t != "" {
			result = append(result, t)
		}
	}
	return result
}

// matchesAnyTarget はテキストブロックがいずれかのアレルゲン文字列を含むか判定します。
func matchesAnyTarget(text string, targets []string) bool {
	for _, t := range targets {
		if strings.Contains(text, t) {
			return true
		}
	}
	return false
}
