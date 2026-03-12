// Command main は開発用にローカルで Cloud Function を起動します。
//
// 使い方:
//
//	go run ./cmd/main.go
//	# その後 http://localhost:8080 にリクエストを送信してください
package main

import (
	"log"
	"os"

	// 関数パッケージをインポートして init() 登録をトリガーする。
	_ "example.com/aller-navi/highlight-pdf"
	"github.com/GoogleCloudPlatform/functions-framework-go/funcframework"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	if err := funcframework.Start(port); err != nil {
		log.Fatalf("funcframework.Start: %v", err)
	}
}
