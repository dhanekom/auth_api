package main

import (
	"auth_api/internal/middleware"
	"net/http"
)

func (app *Configs) routes() http.Handler {
	router := http.NewServeMux()

	router.HandleFunc("POST /auth/register", app.RegisterHandler)
	router.HandleFunc("GET /auth/verify", app.GenerateVerificationCodeHandler)
	router.HandleFunc("POST /auth/verify", app.VerifyUserHandler)
	router.HandleFunc("POST /auth/token", app.TokenHandler)
	router.HandleFunc("POST /auth/resetpassword", app.ResetPasswordHandler)
	router.HandleFunc("POST /auth/verifypassword", app.VerifyPasswordResetHandler)
	router.HandleFunc("GET /auth/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	adminRouter := http.NewServeMux()
	adminRouter.HandleFunc("DELETE /admin/auth/user", app.DeleteUserHandler)

	router.Handle("/admin/", middleware.Admin(app.AdminTokenSecret, adminRouter))

	v1 := http.NewServeMux()
	v1.Handle("/v1/", http.StripPrefix("/v1", router))

	stack := middleware.CreateStack(
		middleware.Logging,
		middleware.Auth([]string{app.UserTokenSecret, app.AdminTokenSecret}),
	)

	return stack(v1)
}
