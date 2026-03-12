// Command main starts the Cloud Function locally for development.
//
// Usage:
//
//	go run ./cmd/main.go
//	# Then send requests to http://localhost:8080
package main

import (
	"log"
	"os"

	// Import the function package to trigger init() registration.
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
