# Git ワークフロー

## ブランチ戦略

- `main` — プロダクション用。直接 push 禁止
- `claude/<feature>-<id>` — Claude による実装ブランチ
- `feature/<name>` — 通常の機能開発ブランチ

## コミットメッセージ

- 英語、動詞始まり (Add / Fix / Update / Remove / Refactor)
- 1行目は 72 文字以内

```
Add allergen highlight feature to PDF handler

- Use Cloud Vision Document Text Detection for PDF parsing
- Add yellow highlight annotations via pdfcpu
- Add name header text to first page
```

## プッシュ前の確認

```bash
# テストが通ることを確認
cd functions/highlight-pdf && go test ./...

# ビルドが通ることを確認
go build ./...

# フォーマット確認
gofmt -l .
```

## プルリクエスト

- `main` へのマージは PR 経由
- レビューが通ってからマージ
- タイトルはコミットメッセージと同様のスタイル

### PR の言語

**PR のタイトル・本文はすべて日本語で記述すること。**

```
タイトル例: PDF ハイライト機能にアレルゲン複数指定を追加

## 概要
アレルゲンを複数同時に指定できるよう機能を拡張した。

## 変更内容
- `target` フィールドをカンマ区切りリストとして解析
- 各アレルゲンに異なる色のハイライトを適用

## テスト
- `go test ./...` がすべて通ることを確認
```
