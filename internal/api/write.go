package api

import (
	"encoding/json"
	"net/http"
)

type errBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

const maxJSONBody = 1 << 20

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, code, msg string) {
	writeJSON(w, status, errBody{Code: code, Message: msg})
}
