package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

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
	r.Route("/remove_user", func(r chi.Router) {
		r.Use()
	})
	r.Route("/login", func(r chi.Router) {
		r.Use(VerifyTypeJSON)
		r.Post("/", loginUser)
	})
	r.Route("/refresh", func(r chi.Router) {
		r.Use(VerifyTypeJSON)
	})
	r.Route("/checkjwt", func(r chi.Router) {
		r.Use(TokenVerify)
		r.Get("/", checkJwt)
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
	WriteJSON(w, result, 200)
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
	WriteJSON(w, country, 200)
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

type tokenResponse struct {
	AccessToken  string
	RefreshToken string
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
	user, err := db.DbService().SelectUserAuth(u.Username)
	if err != nil {
		fmt.Println("Username Failed", err)
		http.Error(w, "Invalid Credentials", 401)
		return
	}

	if user.PasswordHash == "" {
		http.Error(w, "Password Change Needed", 409)
		return
	}

	if user.IsActive == false {
		http.Error(w, "Account Deactivated", 403)
		return
	}

	pw_valid, err := utils.VerifyPassword(user.PasswordHash, u.Password)
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, "Credential Validation Error", 500)
		return
	}
	if !pw_valid {
		http.Error(w, "Invalid Credentials", 401)
		return
	}

	newSessionId, err := db.DbService().NewUserSession(user.Id)
	if err != nil {
		http.Error(w, "New Session Error", 500)
		return
	}

	userClaims := utils.TokenClaims{
		User_id:  user.Id,
		Username: user.Username,
		Is_staff: user.IsStaff,
		Exp:      time.Now().UTC().Add(time.Minute * 15),
	}
	accessToken, err := utils.GenerateAccessToken(&userClaims)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
	userTokens := tokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newSessionId,
	}
	// JWT Needed Here
	WriteJSON(w, userTokens, 201)
}

func checkJwt(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*utils.TokenClaims)
	fmt.Println(user.Username)
	w.WriteHeader(200)
}
