package highlightpdf

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

const maxUploadSize = 20 << 20 // 20MB

// Handler is the Cloud Functions HTTP entry point.
func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("parse form: %v", err))
		return
	}

	// Parse name
	name := r.FormValue("name")
	if name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	// Parse allergens
	allergensJSON := r.FormValue("allergens")
	if allergensJSON == "" {
		writeError(w, http.StatusBadRequest, "allergens is required")
		return
	}
	var allergens []string
	if err := json.Unmarshal([]byte(allergensJSON), &allergens); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("allergens must be a JSON array: %v", err))
		return
	}
	if len(allergens) == 0 {
		writeError(w, http.StatusBadRequest, "allergens must not be empty")
		return
	}

	// Parse PDF file
	file, _, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("file is required: %v", err))
		return
	}
	defer file.Close()

	pdfBytes, err := io.ReadAll(file)
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("read file: %v", err))
		return
	}

	slog.Info("processing request",
		"name", name,
		"allergen_count", len(allergens),
		"pdf_size_bytes", len(pdfBytes),
	)

	// Detect text via Cloud Vision
	pages, err := DetectText(r.Context(), pdfBytes)
	if err != nil {
		slog.Error("vision API error", "err", err)
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("text detection failed: %v", err))
		return
	}

	slog.Info("vision detection complete", "page_count", len(pages))

	// Process PDF: add highlights and name header
	result, err := ProcessPDF(pdfBytes, pages, allergens, name)
	if err != nil {
		slog.Error("PDF processing error", "err", err)
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("PDF processing failed: %v", err))
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", `attachment; filename="highlighted.pdf"`)
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(result); err != nil {
		slog.Error("write response", "err", err)
	}
}

func writeError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(map[string]string{"error": msg}); err != nil {
		slog.Error("write error response", "err", err)
	}
}
