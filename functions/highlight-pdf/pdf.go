package highlightpdf

import (
	"bytes"
	"fmt"
	"log/slog"
	"math"
	"os"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

// highlightRect は PDF ポイント座標系での矩形（左下原点）を表します。
type highlightRect struct{ x, y, w, h float64 }

// ProcessPDF は PDF の target 文字列を含むテキストブロックにハイライトを追加します。
// アノテーションではなくページのコンテンツストリームに直接描画するため、
// LINE 画像プレビューを含むあらゆる PDF レンダラーで表示されます。
func ProcessPDF(pdfBytes []byte, pages []PageInfo, target string) ([]byte, error) {
	dims, err := api.PageDims(bytes.NewReader(pdfBytes), nil)
	if err != nil {
		return nil, fmt.Errorf("get page dims: %w", err)
	}

	targets := splitTargets(target)
	slog.Info("ProcessPDF start",
		"page_count", len(pages),
		"targets", targets,
	)

	// ページ番号 → 矩形リスト のマップを構築する。
	pageRects := make(map[int][]highlightRect)

	for pageIdx, page := range pages {
		if pageIdx >= len(dims) {
			break
		}
		pdfW := dims[pageIdx].Width
		pdfH := dims[pageIdx].Height

		slog.Info("processing page",
			"page", pageIdx+1,
			"image_w", page.Width,
			"image_h", page.Height,
			"pdf_w", pdfW,
			"pdf_h", pdfH,
			"block_count", len(page.Blocks),
		)

		if page.Width == 0 || page.Height == 0 {
			slog.Warn("skipping page: zero image dimensions", "page", pageIdx+1)
			continue
		}
		scaleX := pdfW / float64(page.Width)
		scaleY := pdfH / float64(page.Height)

		pageNum := pageIdx + 1 // pdfcpu はページ番号が 1 始まり

		for _, block := range page.Blocks {
			slog.Debug("checking block", "page", pageNum, "text", block.Text)
			if !matchesAnyTarget(block.Text, targets) {
				continue
			}

			// 画像ピクセル座標（左上原点）を PDF ポイント座標（左下原点）に変換する。
			pdfX1 := block.X1 * scaleX
			pdfY1 := pdfH - (block.Y2 * scaleY) // Y 軸を反転
			pdfX2 := block.X2 * scaleX
			pdfY2 := pdfH - (block.Y1 * scaleY) // Y 軸を反転

			slog.Info("adding highlight",
				"page", pageNum,
				"text", block.Text,
				"rect", fmt.Sprintf("(%.1f,%.1f)-(%.1f,%.1f)", pdfX1, pdfY1, pdfX2, pdfY2),
			)

			pageRects[pageNum] = append(pageRects[pageNum], highlightRect{
				x: pdfX1,
				y: pdfY1,
				w: pdfX2 - pdfX1,
				h: pdfY2 - pdfY1,
			})
		}

		// 複数単語にまたがるターゲット（例：「主食バターロール」）のマッチング。
		// パラグラフ内の連続する単語を結合し、ターゲットと完全一致する場合のみハイライトを追加する。
		// 部分一致（strings.Contains）ではなく完全一致（==）を使用することで、
		// 「牛乳」のような単一単語ターゲットが隣接単語との結合テキストにマッチして
		// 不必要に広い範囲がハイライトされる問題を防ぐ。
		for _, para := range page.Paragraphs {
			words := para.Words
			for windowSize := 2; windowSize <= len(words); windowSize++ {
				for i := 0; i <= len(words)-windowSize; i++ {
					var sb strings.Builder
					for j := i; j < i+windowSize; j++ {
						sb.WriteString(words[j].Text)
					}
					text := sb.String()
					if !exactMatchesAnyTarget(text, targets) {
						continue
					}

					matchedWords := words[i : i+windowSize]

					if wordsOnSameLine(matchedWords) {
						// 同一行: 結合バウンディングボックスで 1 矩形
						wx1 := matchedWords[0].X1
						wy1 := matchedWords[0].Y1
						wx2 := matchedWords[0].X2
						wy2 := matchedWords[0].Y2
						for _, w := range matchedWords[1:] {
							if w.X1 < wx1 {
								wx1 = w.X1
							}
							if w.Y1 < wy1 {
								wy1 = w.Y1
							}
							if w.X2 > wx2 {
								wx2 = w.X2
							}
							if w.Y2 > wy2 {
								wy2 = w.Y2
							}
						}
						pdfX1 := wx1 * scaleX
						pdfY1 := pdfH - (wy2 * scaleY)
						pdfX2 := wx2 * scaleX
						pdfY2 := pdfH - (wy1 * scaleY)
						slog.Info("adding highlight (multi-word)",
							"page", pageNum,
							"text", text,
							"rect", fmt.Sprintf("(%.1f,%.1f)-(%.1f,%.1f)", pdfX1, pdfY1, pdfX2, pdfY2),
						)
						pageRects[pageNum] = append(pageRects[pageNum], highlightRect{
							x: pdfX1,
							y: pdfY1,
							w: pdfX2 - pdfX1,
							h: pdfY2 - pdfY1,
						})
					} else {
						// 複数行またぎ: 単語ごとに個別矩形を追加することで
						// 行間に存在する無関係テキストを誤ってハイライトしない。
						for _, w := range matchedWords {
							pdfX1 := w.X1 * scaleX
							pdfY1 := pdfH - (w.Y2 * scaleY)
							pdfX2 := w.X2 * scaleX
							pdfY2 := pdfH - (w.Y1 * scaleY)
							slog.Info("adding highlight (multi-word, cross-line)",
								"page", pageNum,
								"text", text,
								"word", w.Text,
								"rect", fmt.Sprintf("(%.1f,%.1f)-(%.1f,%.1f)", pdfX1, pdfY1, pdfX2, pdfY2),
							)
							pageRects[pageNum] = append(pageRects[pageNum], highlightRect{
								x: pdfX1,
								y: pdfY1,
								w: pdfX2 - pdfX1,
								h: pdfY2 - pdfY1,
							})
						}
					}
				}
			}
		}
	}

	if len(pageRects) == 0 {
		slog.Warn("no highlights added: no text blocks matched any target")
		return pdfBytes, nil
	}

	// ページ別にハイライト用 PDF を一時ファイルとして作成し、
	// pdfcpu のウォーターマーク機能でコンテンツストリームに直接埋め込む。
	// OnTop=false にすることで、テキストの下（コンテンツストリームの先頭）に描画される。
	watermarksMap := make(map[int][]*model.Watermark)
	var tempFiles []string
	defer func() {
		for _, f := range tempFiles {
			os.Remove(f)
		}
	}()

	for pageNum, rects := range pageRects {
		pdfW := dims[pageNum-1].Width
		pdfH := dims[pageNum-1].Height

		tmpPath, err := writeHighlightPDF(pdfW, pdfH, rects)
		if err != nil {
			return nil, fmt.Errorf("create highlight PDF page %d: %w", pageNum, err)
		}
		tempFiles = append(tempFiles, tmpPath)

		// scalefactor:1 abs でウォーターマーク PDF をそのままのサイズで配置する。
		// 対象ページと同じ MediaBox を持つ PDF を中央揃えで配置すると座標が 1:1 で一致する。
		wm, err := api.PDFWatermark(tmpPath+":1", "scalefactor:1 abs, rot:0", false, false, types.POINTS)
		if err != nil {
			return nil, fmt.Errorf("create watermark page %d: %w", pageNum, err)
		}
		watermarksMap[pageNum] = []*model.Watermark{wm}
	}

	var buf bytes.Buffer
	if err := api.AddWatermarksSliceMap(bytes.NewReader(pdfBytes), &buf, watermarksMap, nil); err != nil {
		return nil, fmt.Errorf("add watermarks: %w", err)
	}

	return buf.Bytes(), nil
}

// writeHighlightPDF はページサイズと矩形リストから黄色矩形を描画する最小限の PDF を生成し、
// 一時ファイルに書き出してそのパスを返します。
func writeHighlightPDF(width, height float64, rects []highlightRect) (string, error) {
	// コンテンツストリーム: 水色（R=0.529 G=0.808 B=0.922）で矩形を塗りつぶす
	var cs bytes.Buffer
	cs.WriteString("q\n0.529 0.808 0.922 rg\n")
	for _, r := range rects {
		fmt.Fprintf(&cs, "%.3f %.3f %.3f %.3f re\n", r.x, r.y, r.w, r.h)
	}
	cs.WriteString("f\nQ\n")
	csData := cs.Bytes()

	// 最小限の PDF を組み立てる（xref オフセットを手動計算）
	var b bytes.Buffer
	b.WriteString("%PDF-1.4\n")

	off1 := b.Len()
	b.WriteString("1 0 obj\n<</Type /Catalog /Pages 2 0 R>>\nendobj\n")

	off2 := b.Len()
	b.WriteString("2 0 obj\n<</Type /Pages /Kids [3 0 R] /Count 1>>\nendobj\n")

	off3 := b.Len()
	fmt.Fprintf(&b, "3 0 obj\n<</Type /Page /Parent 2 0 R /MediaBox [0 0 %.3f %.3f] /Contents 4 0 R /Resources <<>>>>\nendobj\n", width, height)

	off4 := b.Len()
	fmt.Fprintf(&b, "4 0 obj\n<</Length %d>>\nstream\n", len(csData))
	b.Write(csData)
	b.WriteString("endstream\nendobj\n")

	xrefOff := b.Len()
	fmt.Fprintf(&b,
		"xref\n0 5\n0000000000 65535 f \n%010d 00000 n \n%010d 00000 n \n%010d 00000 n \n%010d 00000 n \ntrailer\n<</Size 5 /Root 1 0 R>>\nstartxref\n%d\n%%%%EOF\n",
		off1, off2, off3, off4, xrefOff,
	)

	f, err := os.CreateTemp("", "aller-navi-hl-*.pdf")
	if err != nil {
		return "", err
	}
	defer f.Close()
	if _, err := f.Write(b.Bytes()); err != nil {
		os.Remove(f.Name())
		return "", err
	}
	return f.Name(), nil
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
// 単語レベルのブロックに対して部分一致で使用します。
func matchesAnyTarget(text string, targets []string) bool {
	for _, t := range targets {
		if strings.Contains(text, t) {
			return true
		}
	}
	return false
}

// exactMatchesAnyTarget はテキストがいずれかのターゲット文字列と完全一致するか判定します。
// 複数単語を結合したスパンに対して使用し、部分一致による過剰ハイライトを防ぎます。
func exactMatchesAnyTarget(text string, targets []string) bool {
	for _, t := range targets {
		if text == t {
			return true
		}
	}
	return false
}

// wordsOnSameLine は単語リストがすべて同一視覚行にあるかを判定します。
// 連続する 2 単語間の Y 中心座標の差が平均単語高さ以内であれば同一行とみなします。
func wordsOnSameLine(words []TextBlock) bool {
	for i := 1; i < len(words); i++ {
		prev := words[i-1]
		curr := words[i]
		prevCenter := (prev.Y1 + prev.Y2) / 2
		currCenter := (curr.Y1 + curr.Y2) / 2
		avgHeight := ((prev.Y2 - prev.Y1) + (curr.Y2 - curr.Y1)) / 2
		if math.Abs(prevCenter-currCenter) > avgHeight {
			return false
		}
	}
	return true
}
