package main

import (
	"auth_api/internal/middleware"
	"net/http"
)

func (app *Configs) routes() http.Handler {
	router := http.NewServeMux()

	router.HandleFunc("POST /api/auth/register", app.RegisterHandler)
	router.HandleFunc("GET /api/auth/verify", app.GenerateVerificationCodeHandler)
	router.HandleFunc("POST /api/auth/verify", app.VerifyUserHandler)
	router.HandleFunc("POST /api/auth/token", app.TokenHandler)
	router.HandleFunc("GET /api/auth/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	adminRouter := http.NewServeMux()
	adminRouter.HandleFunc("DELETE /api/auth/user", app.DeleteUserHandler)

	router.Handle("/", middleware.Admin(app.AdminTokenSecret, adminRouter))

	v1 := http.NewServeMux()
	v1.Handle("/v1/", http.StripPrefix("/v1", router))

	stack := middleware.CreateStack(
		middleware.Logging,
		middleware.Auth([]string{app.UserTokenSecret, app.AdminTokenSecret}),
	)

	return stack(v1)
}
