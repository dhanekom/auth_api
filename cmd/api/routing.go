package main

import (
	"net/http"
)

func (app *Configs) routes() http.Handler {
	router := http.NewServeMux()

	router.HandleFunc("POST /api/auth/register", app.RegisterHandler)
	router.HandleFunc("GET /api/auth/verify", app.GenerateVerificationCodeHandler)
	router.HandleFunc("POST /api/auth/verify", app.VerifyUserHandler)
	router.HandleFunc("POST /api/auth/token", app.TokenHandler)
	router.HandleFunc("DELETE /api/auth/user", app.DeleteUserHandler)

	return router
}
