package main

import (
	"auth_api/internal/middleware"
	"net/http"

	"github.com/justinas/alice"
)

func (app *Configs) routes() http.Handler {
	router := http.NewServeMux()

	router.HandleFunc("POST /auth/register", app.RegisterHandler)
	router.HandleFunc("GET /auth/verifyuser", app.GenerateVerificationCodeHandler)
	router.HandleFunc("POST /auth/verifyuser", app.VerifyUserHandler)
	router.HandleFunc("POST /auth/token", app.TokenHandler)
	router.HandleFunc("POST /auth/resetpassword", app.ResetPasswordRequestHandler)
	router.HandleFunc("PUT /auth/resetpassword", app.ResetPasswordHandler)
	router.HandleFunc("POST /auth/updatepassword", app.UpdatePasswordHandler)
	router.HandleFunc("GET /auth/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	adminRouter := http.NewServeMux()
	adminRouter.HandleFunc("DELETE /admin/auth/user", app.DeleteUserHandler)
	adminRouter.HandleFunc("GET /auth/role", app.UserRoleHandler)

	router.Handle("/admin/", middleware.Admin(app.AdminTokenSecret, adminRouter))

	v1 := http.NewServeMux()
	v1.Handle("/v1/", http.StripPrefix("/v1", router))

	// stack := middleware.CreateStack(
	// 	middleware.Logging,
	// 	middleware.Auth([]string{app.UserTokenSecret, app.AdminTokenSecret}),
	// )

	return alice.New(middleware.Logging, middleware.Auth([]string{app.UserTokenSecret, app.AdminTokenSecret}), middleware.RateLimiter).Then(v1)
	//stack(v1)
}
