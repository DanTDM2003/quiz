package main

import (
	"log"
	"net/http"
	"os"

	"quiz/internal/api"
	"quiz/internal/quiz"
)

func main() {
	reg := quiz.SeededRegistry()
	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/quizzes/{quizId}/participants", api.JoinHandler(reg))
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	log.Printf("listening on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}
