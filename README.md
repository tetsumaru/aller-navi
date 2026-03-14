# aller-navi

給食・保育園のメニュー PDF にアレルゲン情報をハイライトして LINE で返す Google Cloud サービス。

## サービス構成

| サービス | 種別 | 説明 |
|---|---|---|
| **highlight-pdf** | Cloud Functions (Go) | PDF にアレルゲンハイライトを追加する HTTP API |
| **register-allergen** | Cloud Functions (Go) | Firestore にアレルゲン情報を登録する HTTP API |
| **linebot** | Cloud Run (Go + Docker) | LINE Messaging API Webhook。PDF を受け取り、ハイライト済み画像を返信する |

## 機能

- LINE にメニュー PDF を送ると、アレルゲン箇所が黄色でハイライトされた画像が返ってくる
- LINE にアレルゲン文字列（例: `卵\n乳\n小麦`）をテキスト送信すると Firestore に登録される
- ハイライト対象文字列は Firestore `users/{user_id}.target` フィールドで管理
- Google Cloud Vision API (Document Text Detection) でテキストと位置情報を検出
- Ghostscript で PDF → JPEG 変換し、LINE に画像メッセージとして返信

## 開発環境

### 必要なもの

- Go 1.26+
- [Google Cloud SDK](https://cloud.google.com/sdk/docs/install)
- Cloud Vision API・Cloud Firestore・Cloud Storage が有効な GCP プロジェクト
- Docker (linebot のローカル実行時)
- Ghostscript (linebot のローカル実行時)

### セットアップ

```bash
# GCP 認証
gcloud auth application-default login

# highlight-pdf の依存関係取得
cd functions/highlight-pdf
go mod tidy

# linebot の依存関係取得
cd functions/linebot
go mod tidy
```

### ローカル起動

```bash
# highlight-pdf
cd functions/highlight-pdf
go run ./cmd/main.go
# → http://localhost:8080 で起動

# linebot
cd functions/linebot
docker build -t linebot .
docker run -p 8080:8080 \
  -e LINE_CHANNEL_SECRET=xxx \
  -e LINE_CHANNEL_ACCESS_TOKEN=xxx \
  -e HIGHLIGHT_PDF_URL=https://... \
  -e REGISTER_ALLERGEN_URL=https://... \
  -e GCS_BUCKET=your-bucket \
  linebot
```

### リクエスト例

```bash
# PDF ハイライト (highlight-pdf)
curl -X POST http://localhost:8080 \
  -F "file=@sample-menu.pdf" \
  -F "user_id=default" \
  --output highlighted.pdf

# アレルゲン登録 (register-allergen)
curl -X POST https://<REGION>-<PROJECT>.cloudfunctions.net/register-allergen \
  -d "user_id=default" \
  -d "target=卵"
```

### テスト

```bash
cd functions/highlight-pdf
go test ./...
```

### デプロイ

デプロイは GitHub Actions (`main` ブランチへの push) で自動実行される。
詳細は `.github/workflows/deploy.yml` を参照。

```bash
# 手動デプロイ（highlight-pdf）
gcloud functions deploy highlight-pdf \
  --runtime go126 \
  --trigger-http \
  --allow-unauthenticated \
  --region asia-northeast1 \
  --source functions/highlight-pdf

# 手動デプロイ（register-allergen）
gcloud functions deploy register-allergen \
  --runtime go126 \
  --trigger-http \
  --allow-unauthenticated \
  --region asia-northeast1 \
  --source functions/highlight-pdf \
  --entry-point RegisterAllergen

# 手動デプロイ（linebot）— Docker ビルド後に Cloud Run へ
```

## API 仕様

詳細は [`rules/api.md`](rules/api.md) を参照。

## アーキテクチャ

詳細は [`rules/architecture.md`](rules/architecture.md) を参照。

## ライセンス

This project is licensed under the Creative Commons Attribution-NonCommercial 4.0 International License.

[![License: CC BY-NC 4.0](https://img.shields.io/badge/License-CC%20BY--NC%204.0-lightgrey.svg)](https://creativecommons.org/licenses/by-nc/4.0/)
