package linebot

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

// callHighlightPDF は highlight-pdf Cloud Function を呼び出してハイライト済み PDF を返します。
func callHighlightPDF(ctx context.Context, pdfBytes []byte, userID string) ([]byte, error) {
	url := os.Getenv("HIGHLIGHT_PDF_URL")
	if url == "" {
		return nil, fmt.Errorf("HIGHLIGHT_PDF_URL が設定されていません")
	}

	var body bytes.Buffer
	w := multipart.NewWriter(&body)

	if err := w.WriteField("user_id", userID); err != nil {
		return nil, fmt.Errorf("user_id フィールド書き込み: %w", err)
	}

	fw, err := w.CreateFormFile("file", "input.pdf")
	if err != nil {
		return nil, fmt.Errorf("form ファイル作成: %w", err)
	}
	if _, err := fw.Write(pdfBytes); err != nil {
		return nil, fmt.Errorf("PDF 書き込み: %w", err)
	}
	w.Close()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &body)
	if err != nil {
		return nil, fmt.Errorf("リクエスト作成: %w", err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	idToken, err := fetchIDToken(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("ID トークン取得: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+idToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP リクエスト: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("highlight-pdf エラー (status=%d): %s", resp.StatusCode, b)
	}

	return io.ReadAll(resp.Body)
}

// fetchIDToken は GCP メタデータサーバーから audience 向けの ID トークンを取得します。
func fetchIDToken(ctx context.Context, audience string) (string, error) {
	metadataURL := "http://metadata.google.internal/computeMetadata/v1/instance/service-accounts/default/identity?audience=" + audience
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, metadataURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Metadata-Flavor", "Google")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("メタデータサーバーエラー (status=%d): %s", resp.StatusCode, b)
	}

	token, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(token), nil
}
