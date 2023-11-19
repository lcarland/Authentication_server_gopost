package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"authapi/db"
	"authapi/utils"
	. "authapi/utils"
)

// routes at 2/
func apiRoutes(r chi.Router) {
	r.Get("/", index)
	r.Route("/{country}", func(r chi.Router) {
		r.Use(CountryCtx)
		r.Get("/", getCountry)
	})
	r.Route("/register", func(r chi.Router) {
		r.Use(VerifyTypeJSON)
		r.Post("/", createUser)
	})
	r.Route("/login", func(r chi.Router) {
		r.Use(VerifyTypeJSON)
		r.Post("/", loginUser)
	})
}

func index(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("country")
	var result interface{}
	var err error
	if code != "" {
		var query *db.Country
		query, err = db.DbService().GetCountry(code)
		result = query
	} else {
		var query *[]db.Country
		query, err = db.DbService().GetAllCountries()
		result = query
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Println(result)
	WriteJSON(w, result)
}

func CountryCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		countryCode := chi.URLParam(r, "country")
		country, err := db.DbService().GetCountry(countryCode)
		if err != nil {
			http.Error(w, err.Error(), 404)
			return
		}
		ctx := context.WithValue(r.Context(), "country", country)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getCountry(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	country := ctx.Value("country").(*db.Country)
	WriteJSON(w, country)
}

func createUser(w http.ResponseWriter, r *http.Request) {
	// enforce maximum decode size
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	var u db.NewUser
	err := dec.Decode(&u)
	if err != nil {
		http.Error(w, err.Error(), 422)
		return
	}

	err = db.DbService().InsertUser(u)
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), 400)
		return
	}
	w.WriteHeader(201)
}

type userCreds struct {
	Username string
	Password string
}

func loginUser(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	var u userCreds
	err := dec.Decode(&u)
	if err != nil {
		http.Error(w, err.Error(), 422)
		return
	}
	id := db.DbService().GetUserId(u.Username)
	if id == 0 {
		fmt.Println("Username Failed")
		http.Error(w, "Invalid Credentials", 401)
		return
	}

	dbHash := db.DbService().SelectUserHash(id)
	if dbHash == "" {
		http.Error(w, "Password Change Needed", 409)
		return
	}

	pw_valid, _ := utils.VerifyPassword(dbHash, u.Password)
	if !pw_valid {
		fmt.Println("Invalid Password")
		http.Error(w, "Invalid Credentials", 401)
		return
	}
	sessionId := db.DbService().NewUserSession(id)
	if sessionId == "" {
		http.Error(w, "New Session Error", 500)
		return
	}
	// JWT Needed Here
	// Set session to Cookies
	WriteJSON(w, map[any]any{"Msg": "Login Succes", "SessionId": sessionId})
}
