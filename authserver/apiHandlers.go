package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"authapi/db"
	. "authapi/utils"
)

// routes at 2/
func apiRoutes(r chi.Router) {
	r.Get("/", index)
}

func index(w http.ResponseWriter, r *http.Request) {
	code := "US"

	query, _ := db.DbService().GetCountry(code)
	fmt.Println(query)
	WriteJSON(w, query)
}
