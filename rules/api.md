# API 仕様

## エンドポイント

```
POST /highlight-pdf
Content-Type: multipart/form-data
```

## リクエストパラメータ

| フィールド | 型 | 必須 | 説明 |
|---|---|---|---|
| `file` | ファイル (application/pdf) | ✅ | ハイライト対象の PDF ファイル |
| `allergens` | JSON 配列 (文字列) | ✅ | ハイライト対象のアレルゲン文字列のリスト |
| `name` | 文字列 | ✅ | PDF の上部に記載する氏名 |

### allergens の例

```json
["卵", "乳", "小麦", "えび"]
```

### curl の例

```bash
curl -X POST https://<REGION>-<PROJECT>.cloudfunctions.net/highlight-pdf \
  -F "file=@menu.pdf" \
  -F 'allergens=["卵","乳"]' \
  -F "name=山田 太郎" \
  --output highlighted.pdf
```

## レスポンス

### 成功 (200 OK)

```
Content-Type: application/pdf
Content-Disposition: attachment; filename="highlighted.pdf"

<PDF バイナリ>
```

- 1ページ目上部に氏名が追記されている
- アレルゲン文字列を含むテキストブロックが黄色でハイライトされている

### エラー

| ステータス | 説明 |
|---|---|
| 400 Bad Request | リクエストパラメータ不正 (ファイル未添付、allergens が不正な JSON 等) |
| 405 Method Not Allowed | POST 以外のメソッド |
| 500 Internal Server Error | Cloud Vision API エラー、PDF 処理エラー |

エラーレスポンスボディ:
```json
{"error": "<エラーメッセージ>"}
```
