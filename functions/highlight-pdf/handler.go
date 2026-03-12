package highlightpdf

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

const maxUploadSize = 20 << 20 // 20MB（アップロード上限サイズ）

// Handler は Cloud Functions の HTTP エントリーポイントです。
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

	// ユーザー ID を取得する
	userID := r.FormValue("user_id")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "user_id is required")
		return
	}

	// PDF ファイルを取得する
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

	// Firestore からハイライト対象文字列を取得する
	target, err := GetUserTarget(r.Context(), userID)
	if err != nil {
		slog.Error("firestore error", "err", err)
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("get user target: %v", err))
		return
	}

	slog.Info("processing request",
		"user_id", userID,
		"target", target,
		"pdf_size_bytes", len(pdfBytes),
	)

	// Cloud Vision API でテキストを検出する
	pages, err := DetectText(r.Context(), pdfBytes)
	if err != nil {
		slog.Error("vision API error", "err", err)
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("text detection failed: %v", err))
		return
	}

	slog.Info("vision detection complete", "page_count", len(pages))

	// PDF を処理する：ハイライトを追加する
	result, err := ProcessPDF(pdfBytes, pages, target)
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
