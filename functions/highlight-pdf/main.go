// Package highlightpdf provides a Cloud Function that highlights allergen-related
// text in PDF files using Google Cloud Vision API.
package highlightpdf

import (
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
)

func init() {
	functions.HTTP("HighlightPDF", Handler)
}
