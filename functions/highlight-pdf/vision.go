package highlightpdf

import (
	"context"
	"fmt"
	"strings"

	vision "cloud.google.com/go/vision/apiv1"
	visionpb "cloud.google.com/go/vision/apiv1/visionpb"
)

// PageInfo holds text blocks detected on a single PDF page.
type PageInfo struct {
	// Width and Height are the rendered image dimensions returned by Vision API (pixels).
	Width  int32
	Height int32
	Blocks []TextBlock
}

// TextBlock represents a paragraph-level text region with its bounding box
// in image pixel coordinates (top-left origin, Y-axis downward).
type TextBlock struct {
	Text string
	X1   float64 // left
	Y1   float64 // top
	X2   float64 // right
	Y2   float64 // bottom
}

// DetectText sends PDF bytes to Cloud Vision API and returns text blocks per page.
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
					GcsSourceOrContent: &visionpb.InputConfig_Content{
						Content: pdfBytes,
					},
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
						text := extractParagraphText(para)
						if text == "" {
							continue
						}
						x1, y1, x2, y2 := polyBounds(para.GetBoundingBox())
						pi.Blocks = append(pi.Blocks, TextBlock{
							Text: text,
							X1:   x1, Y1: y1,
							X2:   x2, Y2: y2,
						})
					}
				}

				pages = append(pages, pi)
			}
		}
	}

	return pages, nil
}

// extractParagraphText concatenates all symbol texts in a paragraph.
func extractParagraphText(para *visionpb.Paragraph) string {
	var sb strings.Builder
	for _, word := range para.GetWords() {
		for _, sym := range word.GetSymbols() {
			sb.WriteString(sym.GetText())
		}
	}
	return sb.String()
}

// polyBounds returns the axis-aligned bounding box of a BoundingPoly.
func polyBounds(poly *visionpb.BoundingPoly) (x1, y1, x2, y2 float64) {
	if poly == nil || len(poly.GetVertices()) == 0 {
		return 0, 0, 0, 0
	}
	verts := poly.GetVertices()
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
