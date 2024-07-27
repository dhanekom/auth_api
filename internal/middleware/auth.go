package middleware

import (
	"auth_api/internal/helpers"
	"auth_api/internal/verify"
	"net/http"
	"strings"
)

func Auth(secrets []string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString := r.Header.Get("Authorization")
			if tokenString == "" {
				helpers.WriteJSON(w, http.StatusUnauthorized, helpers.ErrorResponse("authorization failed"))
				return
			}

			tokenParts := strings.Split(tokenString, " ")
			if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
				helpers.WriteJSON(w, http.StatusUnauthorized, helpers.ErrorResponse("authorization failed"))
				return
			}

			authorized := false
			for _, secret := range secrets {
				if authorized {
					break
				}

				tokenUtils := verify.JWTTokenUtils{}
				tokenUtils.Setup(secret)
				err := tokenUtils.ValidateToken(tokenParts[1])
				if err != nil {
					continue
				}

				authorized = true
			}

			if !authorized {
				helpers.WriteJSON(w, http.StatusUnauthorized, helpers.ErrorResponse("token verification failed"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
