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

// highlightColor is the color applied to allergen-matching text blocks (semi-transparent yellow).
var highlightColor = color.SimpleColor{R: 1.0, G: 0.95, B: 0.0}

// ProcessPDF adds allergen highlights and a name header to the PDF.
// pages must correspond 1:1 with the PDF pages (index 0 = page 1).
func ProcessPDF(pdfBytes []byte, pages []PageInfo, allergens []string, name string) ([]byte, error) {
	// Retrieve PDF page dimensions.
	dims, err := api.PageDimsFile(bytes.NewReader(pdfBytes), nil)
	if err != nil {
		return nil, fmt.Errorf("get page dims: %w", err)
	}

	result := pdfBytes

	// Add highlight annotations for each matched text block.
	for pageIdx, page := range pages {
		if pageIdx >= len(dims) {
			break
		}
		pdfW := dims[pageIdx].Width
		pdfH := dims[pageIdx].Height

		// Avoid division by zero.
		if page.Width == 0 || page.Height == 0 {
			continue
		}
		scaleX := pdfW / float64(page.Width)
		scaleY := pdfH / float64(page.Height)

		pageNum := pageIdx + 1 // pdfcpu uses 1-based page numbers

		for _, block := range page.Blocks {
			if !containsAny(block.Text, allergens) {
				continue
			}

			// Convert image pixel coords (top-left origin) to PDF point coords (bottom-left origin).
			pdfX1 := block.X1 * scaleX
			pdfY1 := pdfH - (block.Y2 * scaleY) // flip Y: image bottom → PDF bottom
			pdfX2 := block.X2 * scaleX
			pdfY2 := pdfH - (block.Y1 * scaleY) // flip Y: image top → PDF top

			rect := types.NewRectangle(pdfX1, pdfY1, pdfX2, pdfY2)

			ann := model.NewSquareAnnotation(
				*rect,
				"",                // contents
				"",                // id
				pageNum,
				nil,               // appearance ref
				&highlightColor,   // fill color (interior)
				nil,               // border color
				0,                 // border width (0 = no border)
			)

			slog.Info("adding highlight",
				"page", pageNum,
				"text", block.Text,
				"rect", fmt.Sprintf("(%.1f,%.1f)-(%.1f,%.1f)", pdfX1, pdfY1, pdfX2, pdfY2),
			)

			var buf bytes.Buffer
			if err := api.AddAnnotations(
				bytes.NewReader(result),
				&buf,
				[]string{fmt.Sprintf("%d", pageNum)},
				ann,
				nil,
			); err != nil {
				return nil, fmt.Errorf("add annotation page %d: %w", pageNum, err)
			}
			result = buf.Bytes()
		}
	}

	// Add name header at the top of page 1.
	result, err = addNameHeader(result, name)
	if err != nil {
		return nil, fmt.Errorf("add name header: %w", err)
	}

	return result, nil
}

// addNameHeader stamps the given name at the top-left of page 1.
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

// containsAny reports whether text contains at least one of the allergen strings.
func containsAny(text string, allergens []string) bool {
	for _, a := range allergens {
		if strings.Contains(text, a) {
			return true
		}
	}
	return false
}
