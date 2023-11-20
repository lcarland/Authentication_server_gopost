package main

import (
	"authapi/utils"
	"context"
	"net/http"
	"strings"
)

func template(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	})
}

func tokenVerify(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeaderString := r.Header["Authorization"]

		headerVal := strings.Split(strings.TrimSpace(authHeaderString[0]), ": ")
		if headerVal[0] != "Bearer" {
			http.Error(w, "Invalid Authorization Method", 401)
			return
		}
		tokenClaims, err := utils.ValidateAccessToken(headerVal[1])
		if err != nil {
			http.Error(w, err.Error(), 401)
			return
		}
		ctx := context.WithValue(r.Context(), "user", tokenClaims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
