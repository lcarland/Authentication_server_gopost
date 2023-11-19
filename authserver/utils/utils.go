package utils

import (
	"encoding/json"
	"net/http"
)

func WriteJSON(w http.ResponseWriter, v any) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(v)
}

func ServerError(w http.ResponseWriter) {
	w.WriteHeader(500)
	w.Write([]byte("Internal Server Error"))
}
