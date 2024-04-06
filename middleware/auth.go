package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/joho/godotenv"
	"github.com/ralphvw/go-template/helpers"
)

func CheckToken(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			helpers.HandleOptions(w, r)
			return
		}

		helpers.EnableCors(w)

		if err := godotenv.Load(); err != nil {
			helpers.LogAction("Error loading env file")
		}

		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			helpers.LogAction("UNAUTHORIZED REQUEST: NO TOKEN PROVIDED")
			http.Error(w, "Unauthorized: No token provided", http.StatusUnauthorized)
			return
		}

		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)

		claims, err := helpers.DecodeToken(tokenString)
		if err != nil {
			helpers.LogAction("INVALID TOKEN" + err.Error())
			http.Error(w, "Unauthorized: Invalid token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "userClaims", claims)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
