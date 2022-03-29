package main

import (
	"io"
	"log"
	"net/http"

	"github.com/joho/godotenv"
)

func main() {
	// .envの読み込み
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	h1 := func(w http.ResponseWriter, _ *http.Request) {
		io.WriteString(w, "Hello from a HandleFunc #1!!!\n")
	}
	h2 := func(w http.ResponseWriter, _ *http.Request) {
		io.WriteString(w, "Hello from a HandleFunc #2!\n")
	}

	http.HandleFunc("/", h1)
	http.HandleFunc("/endpoint", h2)
	http.HandleFunc("/youtube", YoutubeHandler)

	log.Fatal(http.ListenAndServe(":8000", nil))
}
