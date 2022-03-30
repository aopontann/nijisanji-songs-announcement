package main

import (
	"io"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	// .envの読み込み
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// DB接続初期化
	DBInit()

	h1 := func(w http.ResponseWriter, _ *http.Request) {
		io.WriteString(w, "Hello from a HandleFunc #1!!!\n")
	}

	http.HandleFunc("/", h1)
	http.HandleFunc("/youtube", YoutubeHandler)
	// http.HandleFunc("/seed", Seed) // Seed
	// http.HandleFunc("/seedOut", SeedOut) // Seed

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
