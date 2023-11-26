package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
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

// Get user info. Private info given if requested user is self or staff
func getUserInfo(w http.ResponseWriter, r *http.Request) {
	var userInfo any
	var err error

	user := r.Context().Value("user").(*utils.TokenClaims)
	userRequested, err := strconv.Atoi(chi.URLParam(r, "user_id"))
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if user.User_id == userRequested || user.Is_staff {
		userInfo, err = db.DbService().SelectPrivateUserById(user.User_id)
	} else {
		userInfo, err = db.DbService().SelectPublicUser(userRequested)
	}
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	utils.WriteJSON(w, userInfo, 200)
}

// request JSON for login and account delete
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

// main login handler, requires validateUserCreds middleware
func loginUser(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*db.UserAuth)
	newAccess(w, user)
}

// logout user by removing the refresh token for their current client.
// It is up to the client to delete the Access Token.
func logoutUser(w http.ResponseWriter, r *http.Request) {
	var refresh refreshToken

	r.Body = http.MaxBytesReader(w, r.Body, 1048576)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(&refresh)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	err = db.DbService().DeleteSession(refresh.Token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
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

type UserMod struct {
	Username, Password, FirstName, LastName,
	Email, Phone, Country string
	IsSuper, IsStaff bool
}

func modifyUser(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*utils.TokenClaims)

	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	var u UserMod
	err := dec.Decode(&u)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	if u.Username != user.Username || !user.Is_staff {
		http.Error(w, "You cannot change another user's info", http.StatusForbidden)
		return
	}
	// serialize struct into json, and deserialize into map
	var um map[string]interface{}
	us, _ := json.Marshal(&u)
	_ = json.Unmarshal(us, &um)
	err2 := db.DbService().UpdateUserProfile(user.User_id, um)
	if err2 != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// permanently delete user. ValidateUserCreds required
func deleteUserAccount(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*db.UserAuth)
	err := db.DbService().DeleteUser(user.Id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// JWT test endpoint
func checkJwt(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*utils.TokenClaims)
	fmt.Println(user.Username)
	w.WriteHeader(200)
}

// DANGER. Find an alternative and remove
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
