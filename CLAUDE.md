# aller-navi

## 重要: PR作成時の必須ルール

> **PR のタイトル・本文は必ず日本語で記述すること。英語は使用禁止。**
>
> ```
> # 良い例
> タイトル: go mod tidy を自動実行する CI ワークフローを追加
>
> ## 概要
> PR ごとに go.sum が最新化されるよう GitHub Actions を追加した。
>
> ## 変更内容
> - `.github/workflows/go-mod-tidy.yml` を追加
>
> ## テスト
> - ワークフローが正常に実行されることを確認
> ```

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
**Claude はすべての回答・説明を日本語で行うこと。英語での回答は禁止。**
詳細は `rules/language.md` を参照。

## ルール

詳細は `rules/` を参照。

- `rules/architecture.md` — 設計方針
- `rules/api.md` — API 仕様
- `rules/coding-style.md` — コーディング規約
- `rules/git.md` — Git ワークフロー
- `rules/language.md` — 言語ルール（常に日本語で回答すること）
