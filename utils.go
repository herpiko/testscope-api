package main

import (
	"encoding/json"
	"net/http"

	_ "github.com/lib/pq"
)

func respondError(w http.ResponseWriter, code int, message string) {
	respond(w, code, map[string]string{"error": message})
}

func respond(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
