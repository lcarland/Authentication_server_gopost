package main

import (
	"authapi/db"
	"authapi/utils"
	"context"
	"net/http"
	"strings"
)

func template(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	})
}

func TokenVerify(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeaderString := r.Header.Get("Authorization")

		if authHeaderString == "" {
			http.Error(w, "Authorization header is required", 401)
			return
		}

		headerVal := strings.Split(strings.TrimSpace(authHeaderString), " ")
		if len(headerVal) != 2 || headerVal[0] != "Bearer" {
			http.Error(w, "Invalid Authorization Method", 401)
			return
		}

		tokenClaims, err := utils.ValidateAccessToken(headerVal[1])
		if err != nil {
			if err.Error() == "expired" {
				http.Error(w, "Token Expired", 401)
			} else {
				http.Error(w, err.Error(), 401)
				return
			}
		}
		ctx := context.WithValue(r.Context(), "user", tokenClaims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
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
