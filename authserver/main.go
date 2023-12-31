package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"

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
		AllowedOrigins:   ORIGINS,
		AllowedMethods:   METHODS,
		AllowCredentials: true,
	}))

	r.Route("/", apiRoutes)

	port := fmt.Sprintf(":%s", os.Getenv("GO_PORT"))
	fmt.Printf("Listening on: http://localhost%s\n", port)
	http.ListenAndServe(port, r)
}

func apiRoutes(r chi.Router) {
	r.Route("/", func(r chi.Router) {
		r.Route("/{country}", func(r chi.Router) {
			r.Use(CountryCtx)
			r.Get("/", getCountry)
		})
		r.Get("/", index)
	})
	r.Route("/user", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(VerifyTypeJSON)
			r.Post("/", createUser)
		})
		r.Route("/password", func(r chi.Router) {
			r.Use(VerifyTypeJSON)
			r.Post("/", createPasswordToken)
			r.Put("/", changePassword)
		})
		r.Route("/{user_id}", func(r chi.Router) {
			r.Use(TokenRequired)
			r.Get("/", getUserInfo)
			r.Patch("/", modifyUser)

			r.Group(func(r chi.Router) {
				r.Use(VerifyTypeJSON)
				r.Use(validateUserCreds)
				r.Delete("/", deleteUserAccount)
			})
		})
	})
	r.Route("/session", func(r chi.Router) {
		r.Use(VerifyTypeJSON)
		r.Group(func(r chi.Router) {
			r.Use(validateUserCreds)
			r.Post("/", loginUser)
		})
		r.Group(func(r chi.Router) {
			r.Use(TokenRequired)
			r.Post("/refresh", RefreshAccess)
			r.Delete("/", logoutUser)
		})
	})
	r.Route("/checkjwt", func(r chi.Router) {
		r.Use(TokenRequired)
		r.Get("/", checkJwt)
	})
	r.Get("/publickey", getPublicKey)
}

func VerifyTypeJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentHeader := r.Header.Get("Content-Type")
		if contentHeader == "" {
			msg := "Content-Type Header is blank"
			http.Error(w, msg, http.StatusUnsupportedMediaType)
			return
		}
		if contentHeader != MediaTypes["JSON"] {
			msg := "Unsupported Media Type"
			http.Error(w, msg, http.StatusUnsupportedMediaType)
			return
		}
		next.ServeHTTP(w, r)
	})
}
