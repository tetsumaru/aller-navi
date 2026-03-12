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
| `user_id` | 文字列 | ✅ | Firestore `users/{user_id}` のドキュメント ID |

ハイライト対象の文字列は Firestore の `users/{user_id}.target` フィールドから取得します。

### curl の例

```bash
curl -X POST https://<REGION>-<PROJECT>.cloudfunctions.net/highlight-pdf \
  -F "file=@menu.pdf" \
  -F "user_id=abc123" \
  --output highlighted.pdf
```

## Firestore スキーマ

```
users/{user_id}
  target: string  // この文字列を含むテキストブロックをハイライトする
```

## レスポンス

### 成功 (200 OK)

```
Content-Type: application/pdf
Content-Disposition: attachment; filename="highlighted.pdf"

<PDF バイナリ>
```

- `target` 文字列を含むテキストブロックが黄色でハイライトされている

### エラー

| ステータス | 説明 |
|---|---|
| 400 Bad Request | リクエストパラメータ不正 (ファイル未添付、user_id 未指定等) |
| 405 Method Not Allowed | POST 以外のメソッド |
| 500 Internal Server Error | Firestore エラー、Cloud Vision API エラー、PDF 処理エラー |

エラーレスポンスボディ:
```json
{"error": "<エラーメッセージ>"}
```
