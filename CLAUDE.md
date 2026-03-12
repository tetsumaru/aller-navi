# aller-navi

## 概要

給食・保育園のメニュー PDF にアレルゲン情報をハイライトして返す Google Cloud Functions サービス。

## サービス構成

- **functions/highlight-pdf** — PDF にハイライトを追加する HTTP Cloud Function
  - 言語: Go
  - 外部サービス: Google Cloud Vision API (Document Text Detection), Cloud Firestore
  - PDF 操作: pdfcpu (pure Go)
  - ハイライト対象: Firestore `users/{user_id}.target` フィールドの文字列

## 開発環境

- Go 1.21+
- Google Cloud SDK (`gcloud` コマンド)
- Cloud Vision API および Cloud Firestore が有効な GCP プロジェクト

## コマンド

```bash
# テスト
cd functions/highlight-pdf && go test ./...

# ビルド確認
cd functions/highlight-pdf && go build ./...

# ローカル起動
cd functions/highlight-pdf && go run ./cmd/main.go

# デプロイ
gcloud functions deploy highlight-pdf \
  --runtime go121 \
  --trigger-http \
  --allow-unauthenticated \
  --region asia-northeast1
```

## PR・コミュニケーション言語

**PR のタイトルおよび本文は必ず日本語で記述すること。**
詳細フォーマットは `rules/git.md` の「PR の言語」セクションを参照。

## ルール

詳細は `rules/` を参照。

- `rules/architecture.md` — 設計方針
- `rules/api.md` — API 仕様
- `rules/coding-style.md` — コーディング規約
- `rules/git.md` — Git ワークフロー
