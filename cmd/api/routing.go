package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (app *Configs) routes() http.Handler {
	r := gin.Default()

	r.POST("/api/auth/register", app.RegisterHandler)
	r.GET("/api/auth/verify", app.GenerateVerificationCodeHandler)
	r.POST("/api/auth/verify", app.VerifyUserHandler)
	r.POST("/api/auth/token", app.TokenHandler)
	r.DELETE("/api/auth/user", app.DeleteUserHandler)

	return r
}
