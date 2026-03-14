package highlightpdf

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	vision "cloud.google.com/go/vision/v2/apiv1"
	visionpb "cloud.google.com/go/vision/v2/apiv1/visionpb"
)

// ParagraphInfo は Vision API のパラグラフ内の単語リストを保持します。
// 複数単語にまたがるターゲット（例：「主食バターロール」）のマッチングに使用します。
type ParagraphInfo struct {
	Words []TextBlock
}

// PageInfo は PDF の 1 ページ分の検出テキストブロックを保持します。
type PageInfo struct {
	// Width と Height は Vision API が返すレンダリング済み画像の寸法（ピクセル）です。
	Width      int32
	Height     int32
	Blocks     []TextBlock     // 単語レベルのブロック（部分一致マッチング用）
	Paragraphs []ParagraphInfo // パラグラフ構造（複数単語の完全一致マッチング用）
}

// TextBlock は単語レベルのテキスト領域とそのバウンディングボックスを表します。
// 座標は画像ピクセル座標（左上原点、Y 軸下向き）です。
type TextBlock struct {
	Text string
	X1   float64 // 左端
	Y1   float64 // 上端
	X2   float64 // 右端
	Y2   float64 // 下端
}

// DetectText は PDF バイト列を Cloud Vision API に送信し、ページごとのテキストブロックを返します。
func DetectText(ctx context.Context, pdfBytes []byte) ([]PageInfo, error) {
	client, err := vision.NewImageAnnotatorClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("create vision client: %w", err)
	}
	defer client.Close()

	req := &visionpb.BatchAnnotateFilesRequest{
		Requests: []*visionpb.AnnotateFileRequest{
			{
				InputConfig: &visionpb.InputConfig{
					Content:  pdfBytes,
					MimeType: "application/pdf",
				},
				Features: []*visionpb.Feature{
					{Type: visionpb.Feature_DOCUMENT_TEXT_DETECTION},
				},
			},
		},
	}

	resp, err := client.BatchAnnotateFiles(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("batch annotate files: %w", err)
	}

	var pages []PageInfo
	for _, fileResp := range resp.GetResponses() {
		for _, imgResp := range fileResp.GetResponses() {
			if apiErr := imgResp.GetError(); apiErr != nil {
				return nil, fmt.Errorf("vision API page error: %s", apiErr.GetMessage())
			}

			ann := imgResp.GetFullTextAnnotation()
			if ann == nil {
				pages = append(pages, PageInfo{})
				continue
			}

			for _, page := range ann.GetPages() {
				pi := PageInfo{
					Width:  page.GetWidth(),
					Height: page.GetHeight(),
				}

				for _, block := range page.GetBlocks() {
					for _, para := range block.GetParagraphs() {
						var paraWords []TextBlock

						for _, word := range para.GetWords() {
							text := extractWordText(word)
							if text == "" {
								continue
							}
							x1, y1, x2, y2 := polyBounds(word.GetBoundingBox(), pi.Width, pi.Height)
							slog.Debug("detected word",
								"page", len(pages)+1,
								"text", text,
								"bbox", fmt.Sprintf("(%.0f,%.0f)-(%.0f,%.0f)", x1, y1, x2, y2),
							)
							tb := TextBlock{
								Text: text,
								X1:   x1, Y1: y1,
								X2:   x2, Y2: y2,
							}
							pi.Blocks = append(pi.Blocks, tb)
							paraWords = append(paraWords, tb)
						}

						// 複数単語からなるパラグラフはパラグラフ構造として保存する。
						// 「主食バターロール」のように Vision API が複数単語に分割したターゲットを
						// ProcessPDF 内でスライディングウィンドウ完全一致マッチングで検出するため。
						if len(paraWords) > 1 {
							pi.Paragraphs = append(pi.Paragraphs, ParagraphInfo{Words: paraWords})
						}
					}
				}

				slog.Info("vision page detected",
					"page", len(pages)+1,
					"width", pi.Width,
					"height", pi.Height,
					"block_count", len(pi.Blocks),
				)
				pages = append(pages, pi)
			}
		}
	}

	return pages, nil
}

// extractWordText は単語内の全シンボルテキストを連結して返します。
func extractWordText(word *visionpb.Word) string {
	var sb strings.Builder
	for _, sym := range word.GetSymbols() {
		sb.WriteString(sym.GetText())
	}
	return sb.String()
}

// polyBounds は BoundingPoly の軸平行バウンディングボックスを返します。
// vertices が空の場合は normalized_vertices をページ寸法でスケールして使用します。
func polyBounds(poly *visionpb.BoundingPoly, pageW, pageH int32) (x1, y1, x2, y2 float64) {
	if poly == nil {
		return 0, 0, 0, 0
	}

	if verts := poly.GetVertices(); len(verts) > 0 {
		x1 = float64(verts[0].GetX())
		y1 = float64(verts[0].GetY())
		x2 = x1
		y2 = y1
		for _, v := range verts[1:] {
			x := float64(v.GetX())
			y := float64(v.GetY())
			if x < x1 {
				x1 = x
			}
			if y < y1 {
				y1 = y
			}
			if x > x2 {
				x2 = x
			}
			if y > y2 {
				y2 = y
			}
		}
		return x1, y1, x2, y2
	}

	// normalized_vertices にフォールバック（0〜1 の正規化座標をピクセルに変換）
	normVerts := poly.GetNormalizedVertices()
	if len(normVerts) == 0 {
		return 0, 0, 0, 0
	}
	w := float64(pageW)
	h := float64(pageH)
	nx1 := float64(normVerts[0].GetX())
	ny1 := float64(normVerts[0].GetY())
	nx2 := nx1
	ny2 := ny1
	for _, v := range normVerts[1:] {
		x := float64(v.GetX())
		y := float64(v.GetY())
		if x < nx1 {
			nx1 = x
		}
		if y < ny1 {
			ny1 = y
		}
		if x > nx2 {
			nx2 = x
		}
		if y > ny2 {
			ny2 = y
		}
	}
	return nx1 * w, ny1 * h, nx2 * w, ny2 * h
}
