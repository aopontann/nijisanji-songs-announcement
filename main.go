package main

import (
	"io"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	log.Print("starting server...")
	// .envの読み込み
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("defaulting to port %s", port)
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

	// Start HTTP server.
	log.Printf("listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
