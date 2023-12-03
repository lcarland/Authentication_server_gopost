package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"reflect"
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
	uid := db.DbService().GetUserId(u.Username)
	err = db.DbService().UpdateUserLoginTime(uid)
	if err != nil {
		http.Error(w, "Update Time Failed", http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Location", fmt.Sprintf("/user/%d", uid))
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
	valid, err := db.DbService().QueryToken(refresh.Token, claims.User_id, false)
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
	username   string
	first_name string
	last_name  string
	email      string
	phone      string
	country    string
}

func modifyUser(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*utils.TokenClaims)

	userRequested, err := strconv.Atoi(chi.URLParam(r, "user_id"))
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	var u map[string]interface{}
	err = dec.Decode(&u)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	if user.User_id != userRequested && !user.Is_staff {
		http.Error(w, "You cannot change another user's info", http.StatusForbidden)
		return
	}
	// extension
	// checking json -> map against UserMod struct
	if !validateMap(u, UserMod{}) {
		http.Error(w, "Invalid Fields included that do not exist or cannot be modified", http.StatusBadRequest)
		return
	}

	err2 := db.DbService().UpdateUserProfile(user.User_id, u)
	if err2 != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func createPasswordToken(w http.ResponseWriter, r *http.Request) {
	var reqBody struct {
		Email string `db:"email"`
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1048576)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	dec.Decode(&reqBody)

	uid := db.DbService().GetUserIdWithEmail(reqBody.Email)
	newToken, _ := utils.GenerateCryptoString()
	err := db.DbService().NewUserSession(uid, newToken, true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Printf("Reset Token: %s", newToken)

	// This should go out via email
	// w.WriteHeader(201)
	resjson := map[string]string{"reset_token": newToken}
	utils.WriteJSON(w, resjson, 201)
}

func changePassword(w http.ResponseWriter, r *http.Request) {
	var pwChangeReq struct {
		Token    string `json:"token"`
		Username string `json:"username"`
		Password string `json:"password"`
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&pwChangeReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	uid := db.DbService().GetUserId(pwChangeReq.Username)
	valid, err := db.DbService().QueryToken(pwChangeReq.Token, uid, true)
	if err != nil || !valid {
		http.Error(w, "Invalid Token or Username", http.StatusForbidden)
		return
	}
	err = db.DbService().NewUserHashById(uid, pwChangeReq.Password)
	if err != nil {
		fmt.Println("new user hash returned error")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	db.DbService().DeleteSession(pwChangeReq.Token)

	w.WriteHeader(http.StatusAccepted)
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
	w.Write([]byte(fmt.Sprintf("%d", user.User_id)))
	w.WriteHeader(200)
}

// Public Key Endpoint
func getPublicKey(w http.ResponseWriter, r *http.Request) {
	pubkeyFile, err := os.ReadFile(os.Getenv("PUB_KEY"))
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	utils.WriteText(w, string(pubkeyFile), 200)
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

	err := db.DbService().NewUserSession(user.Id, newToken, false)
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
		return
	}
	userTokens := tokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newToken,
	}
	err = db.DbService().UpdateUserLoginTime(user.Id)
	if err != nil {
		http.Error(w, "Update Time Failed", http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Location", fmt.Sprintf("/user/%d", user.Id))
	utils.WriteJSON(w, userTokens, 201)
}

// extends modifyUser
// validateMapKeys checks if the map keys are valid based on the struct fields.
func validateMap(m map[string]any, s any) bool {
	v := reflect.TypeOf(s)
	for key := range m {
		_, fieldFound := v.FieldByName(key)
		if !fieldFound {
			return false
		}
	}
	return true
}
