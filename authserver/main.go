package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"

	"authapi/db"
)

const (
	secret = "u~wNsbYndsuhP2H^ghQuLXCiLUkny$NNa9g-PT7$TWUpbS@S94xPu"
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
		AllowedOrigins:   ORIGINS,
		AllowedMethods:   METHODS,
		AllowCredentials: true,
	}))

	r.Route("/", apiRoutes)

	port := ":3000"
	fmt.Printf("Listening on: http://localhost%s\n", port)
	http.ListenAndServe(port, r)
}
