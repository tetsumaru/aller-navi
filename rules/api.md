# API 仕様

## highlight-pdf

### エンドポイント

```
POST /
Content-Type: multipart/form-data
```

### リクエストパラメータ

| フィールド | 型 | 必須 | 説明 |
|---|---|---|---|
| `file` | ファイル (application/pdf) | ✅ | ハイライト対象の PDF ファイル |
| `user_id` | 文字列 | ✅ | Firestore `users/{user_id}` のドキュメント ID |

ハイライト対象の文字列は Firestore の `users/{user_id}.target` フィールドから取得します。

### curl の例

```bash
curl -X POST https://<REGION>-<PROJECT>.cloudfunctions.net/highlight-pdf \
  -F "file=@menu.pdf" \
  -F "user_id=default" \
  --output highlighted.pdf
```

### レスポンス

#### 成功 (200 OK)

```
Content-Type: application/pdf
Content-Disposition: attachment; filename="highlighted.pdf"

<PDF バイナリ>
```

- `target` 文字列を含むテキストブロックが黄色でハイライトされている

#### エラー

| ステータス | 説明 |
|---|---|
| 400 Bad Request | リクエストパラメータ不正 (ファイル未添付、user_id 未指定等) |
| 405 Method Not Allowed | POST 以外のメソッド |
| 500 Internal Server Error | Firestore エラー、Cloud Vision API エラー、PDF 処理エラー |

エラーレスポンスボディ:
```json
{"error": "<エラーメッセージ>"}
```

---

## register-allergen

### エンドポイント

```
POST /
Content-Type: application/x-www-form-urlencoded
```

### リクエストパラメータ

| フィールド | 型 | 必須 | 説明 |
|---|---|---|---|
| `user_id` | 文字列 | ✅ | Firestore `users/{user_id}` のドキュメント ID |
| `target` | 文字列 | ✅ | ハイライト対象の文字列（複数行可） |

### curl の例

```bash
curl -X POST https://<REGION>-<PROJECT>.cloudfunctions.net/register-allergen \
  -d "user_id=default" \
  -d "target=卵"
```

### レスポンス

#### 成功 (200 OK)

```json
{"status": "ok"}
```

#### エラー

| ステータス | 説明 |
|---|---|
| 400 Bad Request | user_id または target が未指定 |
| 405 Method Not Allowed | POST 以外のメソッド |
| 500 Internal Server Error | Firestore エラー |

---

## linebot (LINE Messaging API Webhook)

### エンドポイント

```
POST /
X-Line-Signature: <LINE署名>
Content-Type: application/json
```

LINE Messaging API からの Webhook イベントを受け取ります。
エンドポイント URL は LINE Developers コンソールに設定します。

### メッセージの処理

| メッセージ種別 | 処理内容 |
|---|---|
| テキストメッセージ | `register-allergen` を呼び出して Firestore に保存し、登録内容を返信 |
| ファイルメッセージ (.pdf) | `highlight-pdf` を呼び出し、ハイライト済み画像を返信（最大 5 枚） |
| ファイルメッセージ (PDF 以外) | 「PDF ファイルを送信してください」と返信 |

### Firestore スキーマ

```
users/{user_id}
  target: string  // この文字列を含むテキストブロックをハイライトする（改行区切り複数行可）
```

linebot が使用する `user_id` は固定値 `"default"` です。
