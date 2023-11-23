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
)

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
	utils.WriteJSON(w, result, 200)
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
	utils.WriteJSON(w, country, 200)
}

func createUser(w http.ResponseWriter, r *http.Request) {
	// enforce maximum decode size
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	var u db.NewUser
	err := dec.Decode(&u)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
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

// request JSON for login
type userCreds struct {
	Username string
	Password string
}

// Response with both tokens.
// For login and token refreshing.
type tokenResponse struct {
	AccessToken  string
	RefreshToken string
}

// request JSON with refresh token
type refreshToken struct {
	Token string `json:"refresh_token"`
}

func loginUser(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	var u userCreds
	err := dec.Decode(&u)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	user, err := db.DbService().SelectUserAuth(u.Username)
	if err != nil {
		fmt.Println("Username Failed", err)
		http.Error(w, "Invalid Credentials", http.StatusUnauthorized)
		return
	}

	if user.PasswordHash == "" {
		http.Error(w, "Password Change Needed", http.StatusConflict)
		return
	}

	pw_valid, err := utils.VerifyPassword(user.PasswordHash, u.Password)
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, "Credential Validation Error", http.StatusInternalServerError)
		return
	}
	if !pw_valid {
		http.Error(w, "Invalid Credentials", http.StatusUnauthorized)
		return
	}
	// handler extension
	newAccess(w, user)
}

func RefreshAccess(w http.ResponseWriter, r *http.Request) {
	claims, err := TokenVerify(r)
	if err != nil {
		if err.Error() != "expired" {
			http.Error(w, "Invalid Token, please login or check headers", http.StatusUnauthorized)
			return
		}
	}

	user, err := db.DbService().SelectUserAuth(claims.Username)
	if err != nil {
		http.Error(w, "Error finding User", 500)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1048576)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	var refresh refreshToken
	err = dec.Decode(&refresh)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	valid, err := db.DbService().QueryToken(refresh.Token, claims.User_id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	if !valid {
		db.DbService().InvalidateAllSessions(claims.User_id)
		http.Error(w, "Login Required", http.StatusUnauthorized)
		return
	}
	err = db.DbService().InvalidateSession(refresh.Token)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// extension
	newAccess(w, user)
}

func checkJwt(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*utils.TokenClaims)
	fmt.Println(user.Username)
	w.WriteHeader(200)
}

func deleteAllUsers(w http.ResponseWriter, r *http.Request) {
	err := db.DbService().DeleteAllUsers()
	if err != nil {
		http.Error(w, err.Error(), 500)
		fmt.Println(err)
		return
	}
	w.WriteHeader(200)
}

//==============================//
// ---- Handler Extensions ---- //
//==============================//

// Extends login and refresh routes due to shared functionality
func newAccess(w http.ResponseWriter, user *db.UserAuth) {
	if !user.IsActive {
		http.Error(w, "Account Deactivated", http.StatusForbidden)
		return
	}
	newToken, _ := utils.GenerateCryptoString()

	err := db.DbService().NewUserSession(user.Id, newToken)
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
		RefreshToken: newToken,
	}
	utils.WriteJSON(w, userTokens, 201)
}
