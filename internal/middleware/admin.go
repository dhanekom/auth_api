package middleware

import (
	"auth_api/internal/helpers"
	"auth_api/internal/verify"
	"net/http"
	"strings"
)

func Admin(secret string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			helpers.WriteJSON(w, http.StatusForbidden, helpers.ErrorResponse("admin access rights required"))
			return
		}

		tokenParts := strings.Split(tokenString, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			helpers.WriteJSON(w, http.StatusForbidden, helpers.ErrorResponse("admin access rights required"))
			return
		}

		tokenUtils := verify.JWTTokenUtils{}
		tokenUtils.Setup(secret)
		err := tokenUtils.ValidateToken(tokenParts[1])
		if err != nil {
			helpers.WriteJSON(w, http.StatusForbidden, helpers.ErrorResponse("admin access rights required"))
			return
		}

		next.ServeHTTP(w, r)
	})
}
