package linebot

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// callRegisterAllergen は register-allergen Cloud Function を呼び出してアレルゲン情報を登録します。
func callRegisterAllergen(ctx context.Context, target, userID string) error {
	registerURL := os.Getenv("REGISTER_ALLERGEN_URL")
	if registerURL == "" {
		return fmt.Errorf("REGISTER_ALLERGEN_URL が設定されていません")
	}

	body := url.Values{}
	body.Set("user_id", userID)
	body.Set("target", target)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, registerURL, strings.NewReader(body.Encode()))
	if err != nil {
		return fmt.Errorf("リクエスト作成: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	idToken, err := fetchIDToken(ctx, registerURL)
	if err != nil {
		return fmt.Errorf("ID トークン取得: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+idToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP リクエスト: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("register-allergen エラー (status=%d): %s", resp.StatusCode, b)
	}

	return nil
}
