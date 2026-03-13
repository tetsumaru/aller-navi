// Cloud Run および ローカル開発用エントリーポイント
package main

import (
	"log"
	"net/http"
	"os"

	linebot "example.com/aller-navi/linebot"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	http.HandleFunc("/", linebot.Handler)
	log.Printf("linebot :%s で起動中\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
