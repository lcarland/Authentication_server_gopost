package main

import (
	"authapi/db"
	"authapi/utils"
	"context"
	"fmt"
	"net/http"
	"strings"
)

func template(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	})
}

func TokenRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenClaims, err := TokenVerify(r)
		if err != nil {
			errtxt := err.Error()
			if errtxt == "header missing" || errtxt == "invalid" {
				http.Error(w, errtxt, 400)
			} else if errtxt == "expired" {
				http.Error(w, errtxt, 401)
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
			http.Error(w, "Access Forbidden", 403)
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
			http.Error(w, "Not Authorized", 403)
			return
		}
		next.ServeHTTP(w, r)
	})
}
