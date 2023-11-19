package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"

	"authapi/config"
	"authapi/db"
)

func main() {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal(err)
	}

	dbService := db.DbService()
	defer dbService.Close()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   config.ORIGINS,
		AllowedMethods:   config.METHODS,
		AllowCredentials: true,
	}))

	r.Route("/", apiRoutes)

	port := ":3000"
	fmt.Printf("Listening on: http://localhost%s\n", port)
	http.ListenAndServe(port, r)
}

var mediaTypes = map[string]string{
	"JSON": "application/json",
	"text": "text/html",
	"form": "multipart/form-data",
}

func VerifyTypeJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentHeader := r.Header.Get("Content-Type")
		if contentHeader == "" {
			msg := "Content-Type Header is blank"
			http.Error(w, msg, http.StatusUnsupportedMediaType)
			return
		}
		if contentHeader != mediaTypes["JSON"] {
			msg := "Unsupported Media Type"
			http.Error(w, msg, http.StatusUnsupportedMediaType)
			return
		}
		next.ServeHTTP(w, r)
	})
}
