# アーキテクチャ方針

## システム全体構成

```
LINE ユーザー
  │
  │  PDF 送信 / テキスト送信
  ▼
linebot (Cloud Run)
  ├─ テキスト受信 → register-allergen (Cloud Functions) → Firestore に保存
  │
  └─ PDF 受信
       │
       ▼
     highlight-pdf (Cloud Functions)
       ├─ 1. Firestore からハイライト対象文字列を取得 (users/{user_id}.target)
       ├─ 2. Cloud Vision API → Document Text Detection
       │       PDF を送信してテキストとバウンディングボックスを取得
       ├─ 3. アレルゲンマッチング
       │       テキストブロック内にアレルゲン文字列が含まれるか判定
       └─ 4. PDF 加工 (pdfcpu)
               マッチしたブロックに黄色ハイライト矩形を追加
       │
       ▼
     ハイライト済み PDF を linebot に返す
       │
       ▼
     Ghostscript で PDF → JPEG 変換
       │
       ▼
     GCS にアップロード
       │
       ▼
  LINE に画像メッセージで返信
```

## サービス詳細

### highlight-pdf (Cloud Functions)

- HTTP トリガー、`functions/highlight-pdf/` ソース
- エントリーポイント: `HighlightPDF`
- リクエスト: `multipart/form-data` (`file`, `user_id`)
- レスポンス: ハイライト済み PDF バイナリ

### register-allergen (Cloud Functions)

- HTTP トリガー、`functions/highlight-pdf/` ソース（同一パッケージ）
- エントリーポイント: `RegisterAllergen`
- リクエスト: `application/x-www-form-urlencoded` (`user_id`, `target`)
- レスポンス: `{"status": "ok"}`

### linebot (Cloud Run)

- Docker コンテナ、Ghostscript を同梱
- LINE Messaging API Webhook (`/`) を受け取る
- 内部で highlight-pdf・register-allergen を ID トークン認証経由で呼び出す
- GCS バケットに JPEG を一時保存し、公開 URL を LINE に送信

## 技術選定

| コンポーネント | 採用技術 | 理由 |
|---|---|---|
| ランタイム | Go 1.26 | コールドスタートが速い、型安全 |
| highlight-pdf プラットフォーム | Cloud Functions (Gen 2) | サーバーレス、スケーラブル |
| linebot プラットフォーム | Cloud Run | Ghostscript 実行のためコンテナが必要 |
| テキスト検出 | Cloud Vision API (DOCUMENT_TEXT_DETECTION) | 日本語対応、PDF 直接入力対応 |
| PDF 操作 | pdfcpu | pure Go (CGO 不要)、Cloud Functions で動作 |
| PDF → 画像変換 | Ghostscript | 高品質変換、linebot コンテナに同梱 |
| 画像ストレージ | Cloud Storage | LINE への URL 配信のため |
| アレルゲン管理 | Cloud Firestore | ユーザーごとのアレルゲン設定を永続化 |
| サービス間認証 | GCP メタデータサーバー ID トークン | Cloud Functions への認証呼び出し |

## 座標系の変換

Cloud Vision API は PDF の各ページを画像としてレンダリングし、**画像ピクセル座標** (左上原点、Y軸下向き) でバウンディングボックスを返す。
一方 pdfcpu は **PDF ポイント座標** (左下原点、Y軸上向き) を使用する。

```
変換式:
  scaleX = pdf_page_width  / image_width
  scaleY = pdf_page_height / image_height
  pdf_x1 = image_x1 * scaleX
  pdf_y1 = pdf_page_height - (image_y2 * scaleY)  ← Y軸反転
  pdf_x2 = image_x2 * scaleX
  pdf_y2 = pdf_page_height - (image_y1 * scaleY)  ← Y軸反転
```

## 制約事項

- PDF サイズ上限: Vision API のインライン入力上限 (20MB)
- ページ数: Vision API 同期 API は 5 ページまで (超える場合は AsyncBatchAnnotateFiles を使用)
- LINE 返信画像数: LINE Messaging API の制限により最大 5 件
- ハイライト精度: Vision API のテキスト検出精度に依存

## 環境変数

### highlight-pdf / register-allergen

| 変数名 | 説明 |
|---|---|
| `FIRESTORE_DATABASE_ID` | Firestore データベース ID（省略時は `(default)`）|

### linebot

| 変数名 | 説明 |
|---|---|
| `LINE_CHANNEL_SECRET` | LINE チャンネルシークレット |
| `LINE_CHANNEL_ACCESS_TOKEN` | LINE チャンネルアクセストークン |
| `HIGHLIGHT_PDF_URL` | highlight-pdf 関数の URL |
| `REGISTER_ALLERGEN_URL` | register-allergen 関数の URL |
| `GCS_BUCKET` | 画像アップロード先 GCS バケット名 |
