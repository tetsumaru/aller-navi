# コーディング規約 (Go)

## 基本方針

- `gofmt` / `goimports` を使用してフォーマット統一
- Go の公式スタイルガイドに従う: https://go.dev/doc/effective_go

## エラーハンドリング

```go
// NG: エラーを無視しない
result, _ := doSomething()

// OK: エラーをラップしてコンテキストを付与
result, err := doSomething()
if err != nil {
    return nil, fmt.Errorf("doSomething: %w", err)
}
```

## ロギング

- `log/slog` (Go 1.21+) を使用
- 構造化ログで出力する

```go
slog.Info("processing PDF", "page_count", len(pages), "allergen_count", len(allergens))
slog.Error("vision API failed", "err", err)
```

## パッケージ構成

### highlight-pdf (`functions/highlight-pdf/`)

1 つの Cloud Functions パッケージに複数のエントリーポイントを持つ。

- `main.go` — 関数の登録 (`init()`) — `HighlightPDF`, `RegisterAllergen` を登録
- `handler.go` — `HighlightPDF` の HTTP リクエスト/レスポンス処理
- `register_handler.go` — `RegisterAllergen` の HTTP リクエスト/レスポンス処理
- `vision.go` — Cloud Vision API クライアント
- `pdf.go` — PDF 操作 (ハイライト)
- `firestore.go` — Firestore クライアント (`GetUserTarget`, `SetUserTarget`)
- `cmd/main.go` — ローカル起動用エントリーポイント

### linebot (`functions/linebot/`)

Cloud Run 上で動作する LINE Webhook サービス。

- `main.go` — 関数の登録 (`init()`)
- `handler.go` — LINE Webhook イベント処理
- `highlight.go` — highlight-pdf 関数の呼び出し
- `register.go` — register-allergen 関数の呼び出し
- `convert.go` — PDF → JPEG 変換 (Ghostscript)
- `storage.go` — GCS への画像アップロード
- `cmd/main.go` — ローカル起動用エントリーポイント

## HTTP ハンドラ

```go
// エラーレスポンスはJSONで返す
func writeError(w http.ResponseWriter, code int, msg string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
```

## テスト

- 各パッケージに `*_test.go` を作成
- テーブル駆動テストを使用
- 外部 API 依存はインターフェースで抽象化してモック化可能にする

```go
type TextDetector interface {
    DetectText(ctx context.Context, pdfBytes []byte) ([]PageInfo, error)
}
```
