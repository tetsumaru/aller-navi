package linebot

import (
	"context"
	"fmt"
	"os"

	"cloud.google.com/go/storage"
	"github.com/google/uuid"
)

// uploadImages は JPEG 画像を GCS にアップロードして公開 URL のスライスを返します。
func uploadImages(ctx context.Context, images [][]byte) ([]string, error) {
	bucket := os.Getenv("GCS_BUCKET")
	if bucket == "" {
		return nil, fmt.Errorf("GCS_BUCKET が設定されていません")
	}

	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("GCS クライアント作成: %w", err)
	}
	defer client.Close()

	urls := make([]string, 0, len(images))
	prefix := uuid.New().String()

	for i, img := range images {
		objectName := fmt.Sprintf("linebot/%s/page-%02d.jpg", prefix, i+1)

		obj := client.Bucket(bucket).Object(objectName)
		w := obj.NewWriter(ctx)
		w.ContentType = "image/jpeg"
		w.CacheControl = "no-cache"

		if _, err := w.Write(img); err != nil {
			w.Close()
			return nil, fmt.Errorf("GCS 書き込み (page %d): %w", i+1, err)
		}
		if err := w.Close(); err != nil {
			return nil, fmt.Errorf("GCS 書き込み完了 (page %d): %w", i+1, err)
		}

		url := fmt.Sprintf("https://storage.googleapis.com/%s/%s", bucket, objectName)
		urls = append(urls, url)
	}

	return urls, nil
}
