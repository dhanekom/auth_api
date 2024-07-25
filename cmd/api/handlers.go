package main

import (
	"auth_api/internal/models"
	"auth_api/internal/validator"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// RegisterHandler create a user account with a hashed password
func (app *Configs) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type signupValidator struct {
		Email    string
		Password string
		validator.Validator
	}

	if err := app.readJSON(w, r, &body); err != nil {
		app.writeJSON(w, http.StatusBadRequest, ErrorResponse("unable to parse json body"))
		return
	}

	sv := signupValidator{
		Email:    body.Email,
		Password: body.Password,
	}

	// validate data
	sv.CheckRequired(sv.Email, "email")
	sv.CheckRequired(sv.Password, "password")
	sv.CheckValue(validator.IsEmail(sv.Email), "email", "valid email required")

	if !sv.Valid() {
		app.writeJSON(w, http.StatusBadRequest, ErrorResponse(sv.Error()))
		return
	}

	// check if account already exists in DB
	users, err := app.DB.GetUsers(r.Context(), sv.Email)
	if err != nil {
		app.writeJSON(w, http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}

	if len(users) > 0 {
		app.writeJSON(w, http.StatusBadRequest, ErrorResponse("user already exists"))
		return
	}

	// Hash and salt password
	hashedPasswordBytes, err := app.PasswordEncryptor.GenerateHashedPassword(sv.Password)
	if err != nil {
		app.writeJSON(w, http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}

	// Create user
	user := &models.User{
		Email:      sv.Email,
		Password:   string(hashedPasswordBytes),
		IsVerified: false,
	}

	err = app.DB.CreateUser(r.Context(), user)
	if err != nil {
		app.writeJSON(w, http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}

	app.writeJSON(w, http.StatusOK, SuccessResponse(gin.H{"message": "successfully created user"}))
}

// GenerateVerificationCodeHandler generates an verification code for an unverified user
func (app *Configs) GenerateVerificationCodeHandler(w http.ResponseWriter, r *http.Request) {
	var requestBody struct {
		Email string `json:"email"`
		validator.Validator
	}

	if err := app.readJSON(w, r, &requestBody); err != nil {
		app.writeJSON(w, http.StatusBadRequest, ErrorResponse("unable to parse json body"))
		return
	}

	requestBody.CheckRequired(requestBody.Email, "email")
	requestBody.CheckValue(validator.IsEmail(requestBody.Email), "email", "valid email required")
	if !requestBody.Valid() {
		app.writeJSON(w, http.StatusBadRequest, ErrorResponse(requestBody.Error()))
		return
	}

	user, err := app.DB.GetUser(r.Context(), requestBody.Email)
	if errors.Is(err, sql.ErrNoRows) {
		app.writeJSON(w, http.StatusBadRequest, ErrorResponse("user does not exist"))
		return
	}

	if user.IsVerified {
		app.writeJSON(w, http.StatusBadRequest, ErrorResponse("user already verified"))
		return
	}

	if err != nil {
		app.writeJSON(w, http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}

	code, err := app.Verifier.GenerateVerificationCode()
	if err != nil {
		app.writeJSON(w, http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	verification := models.Verification{
		Email:             requestBody.Email,
		VerificationCode:  code,
		ExpiresAt:         time.Now().Add(time.Hour * 24),
		AttemptsRemaining: app.Verifier.MaxRetries(),
	}

	if err := app.DB.InsertOrUpdateVerification(r.Context(), verification); err != nil {
		app.writeJSON(w, http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	var responseBody struct {
		VerificationCode string `json:"verification_code"`
	}

	responseBody.VerificationCode = code

	app.writeJSON(w, http.StatusOK, SuccessResponse(responseBody))
}

// VerifyUserHandler is used to verify a user account given a valid verification code (provided by GenerateVerificationCodeHandler)
func (app *Configs) VerifyUserHandler(w http.ResponseWriter, r *http.Request) {
	var requestBody struct {
		Email            string `json:"email"`
		VerificationCode string `json:"verification_code"`
		validator.Validator
	}

	if err := app.readJSON(w, r, &requestBody); err != nil {
		app.writeJSON(w, http.StatusBadRequest, ErrorResponse("unable to parse json body"))
		return
	}

	requestBody.CheckRequired(requestBody.Email, "email")
	requestBody.CheckValue(validator.IsEmail(requestBody.Email), "email", "valid email required")
	requestBody.CheckRequired(requestBody.VerificationCode, "verification_code")

	if !requestBody.Valid() {
		app.writeJSON(w, http.StatusBadRequest, ErrorResponse(requestBody.Error()))
		return
	}

	user, err := app.DB.GetUser(r.Context(), requestBody.Email)
	if errors.Is(err, sql.ErrNoRows) {
		app.writeJSON(w, http.StatusBadRequest, ErrorResponse("user does not exist"))
		return
	}

	if err != nil {
		app.writeJSON(w, http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	verification, err := app.DB.GetVerification(r.Context(), requestBody.Email)
	if errors.Is(err, sql.ErrNoRows) {
		app.writeJSON(w, http.StatusInternalServerError, ErrorResponse(fmt.Sprintf("no verification data found for user %s", requestBody.Email)))
		return
	}

	if err != nil {
		app.writeJSON(w, http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	if verification.ExpiresAt.Before(time.Now()) || verification.AttemptsRemaining <= 0 {
		err := app.DB.DeleteVerification(r.Context(), requestBody.Email)
		if err != nil {
			app.Logger.Error(err.Error())
		}

		app.writeJSON(w, http.StatusBadRequest, ErrorResponse("verification code has expired"))
		return
	}

	if requestBody.VerificationCode != verification.VerificationCode {
		if verification.AttemptsRemaining >= 0 {
			if err := app.DB.InsertOrUpdateVerification(r.Context(), *verification); err != nil {
				app.writeJSON(w, http.StatusInternalServerError, ErrorResponse(err.Error()))
				return
			}
		}

		app.writeJSON(w, http.StatusBadRequest, ErrorResponse("invalid verification code"))
		return
	}

	user.IsVerified = true
	if err := app.DB.UpdateUser(r.Context(), *user); err != nil {
		app.writeJSON(w, http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	if err := app.DB.DeleteVerification(r.Context(), user.Email); err != nil {
		app.Logger.Error(err.Error())
	}

	app.writeJSON(w, http.StatusOK, SuccessResponse(nil))
}

// TokenHandler verifies a user's email and password and returns a JWT (JSON Web Token) if valid credentials were provided
func (app *Configs) TokenHandler(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		validator.Validator
	}

	if err := app.readJSON(w, r, &body); err != nil {
		app.writeJSON(w, http.StatusBadRequest, ErrorResponse("unable to parse json body"))
		return
	}

	body.CheckRequired(body.Email, "email")
	body.CheckRequired(body.Password, "password")
	body.CheckValue(validator.IsEmail(body.Email), "email", "valid email required")
	if !body.Valid() {
		app.writeJSON(w, http.StatusBadRequest, ErrorResponse(body.Error()))
		return
	}

	user, err := app.DB.GetUser(r.Context(), body.Email)
	if errors.Is(err, sql.ErrNoRows) {
		app.writeJSON(w, http.StatusUnauthorized, ErrorResponse("invalid email or password"))
		return
	}

	if err != nil {
		app.writeJSON(w, http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}

	if !user.IsVerified {
		app.writeJSON(w, http.StatusUnauthorized, ErrorResponse("user not verified"))
		return
	}

	err = app.PasswordEncryptor.CompareHashAndPassword([]byte(user.Password), []byte(body.Password))
	if err != nil {
		app.writeJSON(w, http.StatusUnauthorized, ErrorResponse("invalid email or password"))
		return
	}

	tokenString, err := app.TokenGenerator.GenerateToken(user.UserID, time.Now().Add(time.Hour*24).Unix())
	if err != nil {
		app.writeJSON(w, http.StatusBadRequest, ErrorResponse("auth token generation failed"))
		return
	}

	app.writeJSON(w, http.StatusOK, SuccessResponse(gin.H{"token": tokenString}))
}

// DeleteUserHandler takes an email address and deletes the related user and verification data
func (app *Configs) DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email string `json:"email"`
		validator.Validator
	}

	if err := app.readJSON(w, r, &body); err != nil {
		app.writeJSON(w, http.StatusBadRequest, ErrorResponse("unable to parse json body"))
		return
	}

	body.CheckRequired(body.Email, "email")
	body.CheckValue(validator.IsEmail(body.Email), "email", "valid email required")
	if !body.Valid() {
		app.writeJSON(w, http.StatusBadRequest, ErrorResponse(body.Error()))
		return
	}

	recordsDeleted, err := app.DB.DeleteUser(r.Context(), body.Email)
	if err != nil {
		app.writeJSON(w, http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	if !recordsDeleted {
		app.writeJSON(w, http.StatusOK, SuccessResponse(gin.H{"message": "user not found"}))
		return
	}

	app.writeJSON(w, http.StatusOK, SuccessResponse(nil))
}
