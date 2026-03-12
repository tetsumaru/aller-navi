# アーキテクチャ方針

## サービス概要

```
クライアント
  │  multipart/form-data (file, allergens, name)
  ▼
Cloud Functions (Go, HTTP トリガー)
  ├─ 1. リクエスト解析
  ├─ 2. Cloud Vision API → Document Text Detection
  │       PDF を送信してテキストとバウンディングボックスを取得
  ├─ 3. アレルゲンマッチング
  │       テキストブロック内にアレルゲン文字列が含まれるか判定
  ├─ 4. PDF 加工 (pdfcpu)
  │       ├─ マッチしたブロックに黄色ハイライト矩形を追加
  │       └─ 1ページ目上部に氏名テキストを挿入
  └─ 5. 加工済み PDF をレスポンスとして返却
```

## 技術選定

| コンポーネント | 採用技術 | 理由 |
|---|---|---|
| ランタイム | Go 1.21 | コールドスタートが速い、型安全、ユーザー希望 |
| プラットフォーム | Cloud Functions (Gen 2) | サーバーレス、スケーラブル |
| テキスト検出 | Cloud Vision API (DOCUMENT_TEXT_DETECTION) | 日本語対応、PDF直接入力対応 |
| PDF 操作 | pdfcpu | pure Go (CGO不要)、Cloud Functions で動作 |

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
- ハイライト精度: Vision API のテキスト検出精度に依存
