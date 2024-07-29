package middleware

import (
	"net/http"
)

func RateLimiter(next http.Handler) http.Handler {
	// limiter := rate.NewLimiter(2, 4)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// if !limiter.Allow() {
		// 	helpers.WriteJSON(w, http.StatusTooManyRequests, helpers.ErrorResponse("The API is at capacity, try again later."))
		// 	return
		// }

		next.ServeHTTP(w, r)
	})
}
