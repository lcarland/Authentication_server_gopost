package main

import (
	"authapi/db"
	"authapi/utils"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func TokenRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenClaims, err := TokenVerify(r)
		if err != nil {
			errtxt := err.Error()
			if errtxt == "header missing" || errtxt == "invalid" {
				http.Error(w, errtxt, 400)
			} else if errtxt == "expired" {
				http.Error(w, errtxt, http.StatusUnauthorized)
			} else {
				http.Error(w, err.Error(), 500)
			}
			return
		}
		ctx := context.WithValue(r.Context(), "user", tokenClaims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func TokenVerify(r *http.Request) (*utils.TokenClaims, error) {
	authHeaderString := r.Header.Get("Authorization")

	if authHeaderString == "" {
		return nil, fmt.Errorf("header missing")
	}

	headerVal := strings.Split(strings.TrimSpace(authHeaderString), " ")
	if len(headerVal) != 2 || headerVal[0] != "Bearer" {
		return nil, fmt.Errorf("invalid")
	}

	tokenClaims, err := utils.ValidateAccessToken(headerVal[1])
	if err != nil {
		return tokenClaims, err
	}
	return tokenClaims, nil
}

// Staff and Superuser checks can be used seperate from each other, but both rely on TokenVerify first

// Staff permission check. Placed after TokenVerify
func StaffRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userClaim := r.Context().Value("user").(*utils.TokenClaims)
		if !userClaim.Is_staff {
			http.Error(w, "Access Forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Verify user is active superuser against the database every request.
// This middleware function is intended to be placed after TokenVerify in routes.
func SuperUserVerify(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userClaim := r.Context().Value("user").(*utils.TokenClaims)
		user, _ := db.DbService().SelectUserAuth(userClaim.Username)
		if !user.IsActive || !user.IsSuperuser {
			http.Error(w, "Not Authorized", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Username and Login handler for Logining in user and deleting user
func validateUserCreds(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		ctx := context.WithValue(r.Context(), "user", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
