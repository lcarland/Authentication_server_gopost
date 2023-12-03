package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func WriteJSON(w http.ResponseWriter, v any, status int) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

func WriteText(w http.ResponseWriter, v any, status int) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, v)
	w.WriteHeader(status)
}
