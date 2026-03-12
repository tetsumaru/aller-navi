// Package highlightpdf は、Google Cloud Vision API を使用して
// PDF ファイル内のアレルゲン関連テキストをハイライトする Cloud Function を提供します。
package highlightpdf

import (
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
)

func init() {
	functions.HTTP("HighlightPDF", Handler)
}
