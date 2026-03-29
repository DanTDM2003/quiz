package main

import (
	"log"
	"net/http"
	"os"

	"quiz/internal/api"
	"quiz/internal/quiz"
	"quiz/internal/realtime"
)

func main() {
	reg := quiz.SeededRegistry()
	hub := realtime.NewHub()
	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/quizzes/{quizId}/participants", api.JoinHandler(reg))
	mux.HandleFunc("POST /v1/quizzes/{quizId}/answers", api.SubmitAnswerHandler(reg, hub))
	mux.HandleFunc("GET /v1/quizzes/{quizId}/leaderboard", api.LeaderboardHandler(reg))
	mux.HandleFunc("GET /v1/quizzes/{quizId}/stream", api.StreamHandler(reg, hub))
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("listening on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}
