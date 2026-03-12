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

// highlightColor はアレルゲンに一致するテキストブロックに適用する色（黄色）です。
var highlightColor = color.SimpleColor{R: 1.0, G: 0.95, B: 0.0}

// ProcessPDF は PDF にアレルゲンのハイライトと氏名ヘッダーを追加します。
// pages は PDF のページと 1:1 対応している必要があります（インデックス 0 = 1 ページ目）。
func ProcessPDF(pdfBytes []byte, pages []PageInfo, allergens []string, name string) ([]byte, error) {
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

		for _, block := range page.Blocks {
			if !containsAny(block.Text, allergens) {
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
				0,              // apObjNr
				"", "",         // contents, id
				"",             // modDate
				0,              // AnnotationFlags
				highlightColor, // fill color
				"",             // title
				nil,            // popupIndRef
				0.5,            // ca (opacity)
				0, 0,           // borderRadX, borderRadY
				nil,            // BorderStyle
			)

			slog.Info("adding highlight",
				"page", pageNum,
				"text", block.Text,
				"rect", fmt.Sprintf("(%.1f,%.1f)-(%.1f,%.1f)", pdfX1, pdfY1, pdfX2, pdfY2),
			)

			annotationsMap[pageNum] = append(annotationsMap[pageNum], ann)
		}
	}

	// 全アノテーションを一括で適用する。
	result := pdfBytes
	if len(annotationsMap) > 0 {
		var buf bytes.Buffer
		if err := api.AddAnnotationsMap(bytes.NewReader(pdfBytes), &buf, annotationsMap, nil); err != nil {
			return nil, fmt.Errorf("add annotations: %w", err)
		}
		result = buf.Bytes()
	}

	// 1 ページ目の上部に氏名ヘッダーを追加する。
	result, err = addNameHeader(result, name)
	if err != nil {
		return nil, fmt.Errorf("add name header: %w", err)
	}

	return result, nil
}

// addNameHeader は指定した氏名を 1 ページ目の左上に押印します。
func addNameHeader(pdfBytes []byte, name string) ([]byte, error) {
	wm, err := api.TextWatermark(
		name,
		"fontsize:18, pos:tl, offset:15 -15, fillc:#000000, scale:1 abs, rot:0",
		true,  // onTop
		false, // update
		types.POINTS,
	)
	if err != nil {
		return nil, fmt.Errorf("create text watermark: %w", err)
	}

	var buf bytes.Buffer
	if err := api.AddWatermarks(
		bytes.NewReader(pdfBytes),
		&buf,
		[]string{"1"}, // page 1 only
		wm,
		nil,
	); err != nil {
		return nil, fmt.Errorf("add watermark: %w", err)
	}

	return buf.Bytes(), nil
}

// containsAny はテキストにアレルゲン文字列が少なくとも 1 つ含まれているかを返します。
func containsAny(text string, allergens []string) bool {
	for _, a := range allergens {
		if strings.Contains(text, a) {
			return true
		}
	}
	return false
}
