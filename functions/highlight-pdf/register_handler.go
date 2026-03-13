package highlightpdf

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
)

// RegisterAllergenHandler はアレルゲン情報を Firestore に登録する HTTP エントリーポイントです。
func RegisterAllergenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if err := r.ParseForm(); err != nil {
		writeError(w, http.StatusBadRequest, "parse form failed")
		return
	}

	userID := r.FormValue("user_id")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "user_id is required")
		return
	}

	target := strings.TrimSpace(r.FormValue("target"))
	if target == "" {
		writeError(w, http.StatusBadRequest, "target is required")
		return
	}

	if err := SetUserTarget(r.Context(), userID, target); err != nil {
		slog.Error("firestore error", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to save allergen info")
		return
	}

	slog.Info("allergen registered", "user_id", userID, "target", target)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
		slog.Error("write response", "err", err)
	}
}
