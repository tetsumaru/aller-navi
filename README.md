# aller-navi

給食・保育園のメニュー PDF にアレルゲン情報をハイライトして返す Google Cloud Functions サービス。

## 機能

- PDF ファイルと指定したアレルゲン文字列（卵・乳・小麦など）を受け取る
- Google Cloud Vision API でテキストと位置情報を検出
- アレルゲン文字列を含むテキストブロックを黄色でハイライト
- PDF の 1 ページ目上部に氏名を追加
- 加工済み PDF を返却

## 開発環境

### 必要なもの

- Go 1.21+
- [Google Cloud SDK](https://cloud.google.com/sdk/docs/install)
- Cloud Vision API が有効な GCP プロジェクト

### セットアップ

```bash
# GCP 認証
gcloud auth application-default login

# 依存関係の取得
cd functions/highlight-pdf
go mod tidy
```

### ローカル起動

```bash
cd functions/highlight-pdf
go run ./cmd/main.go
# → http://localhost:8080 で起動
```

### リクエスト例

```bash
curl -X POST http://localhost:8080 \
  -F "file=@sample-menu.pdf" \
  -F 'allergens=["卵","乳","小麦"]' \
  -F "name=山田 太郎" \
  --output highlighted.pdf
```

### テスト

```bash
cd functions/highlight-pdf
go test ./...
```

### デプロイ

```bash
gcloud functions deploy highlight-pdf \
  --runtime go121 \
  --trigger-http \
  --allow-unauthenticated \
  --region asia-northeast1 \
  --source functions/highlight-pdf
```

## API 仕様

詳細は [`rules/api.md`](rules/api.md) を参照。

## ライセンス

This project is licensed under the Creative Commons Attribution-NonCommercial 4.0 International License.

[![License: CC BY-NC 4.0](https://img.shields.io/badge/License-CC%20BY--NC%204.0-lightgrey.svg)](https://creativecommons.org/licenses/by-nc/4.0/)
