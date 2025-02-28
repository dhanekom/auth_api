package main

import (
	"auth_api/internal/helpers"
	"auth_api/internal/models"
	"auth_api/internal/validator"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// RegisterHandler create a user account with a hashed password
func (app *Configs) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var requestBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Role     string `json:"role"`
		validator.Validator
	}

	if err := helpers.ReadJSON(w, r, &requestBody); err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, helpers.ErrorResponse("unable to parse json body"))
		return
	}

	// validate data
	requestBody.CheckRequired(requestBody.Email, "email")
	requestBody.CheckRequired(requestBody.Password, "password")
	requestBody.CheckRequired(requestBody.Role, "role")
	requestBody.CheckValue(validator.IsEmail(requestBody.Email), "email", "valid email required")

	if !requestBody.Valid() {
		helpers.WriteJSON(w, http.StatusBadRequest, helpers.ErrorResponse(requestBody.Error()))
		return
	}

	// check if account already exists in DB
	users, err := app.DB.GetUsers(r.Context(), requestBody.Email)
	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, helpers.ErrorResponse(err.Error()))
		return
	}

	if len(users) > 0 {
		helpers.WriteJSON(w, http.StatusBadRequest, helpers.ErrorResponse("user already exists"))
		return
	}

	// Hash and salt password
	hashedPasswordBytes, err := app.PasswordEncryptor.GenerateHashedPassword(requestBody.Password)
	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, helpers.ErrorResponse(err.Error()))
		return
	}

	// Create user
	user := &models.User{
		UserID:   uuid.New().String(),
		Email:    requestBody.Email,
		Password: string(hashedPasswordBytes),
		Status:   models.UserStatusVerifyAccount,
		Role:     requestBody.Role,
	}

	err = app.DB.CreateUser(r.Context(), user)
	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, helpers.ErrorResponse(err.Error()))
		return
	}

	helpers.WriteJSON(w, http.StatusOK, helpers.SuccessResponse(map[string]any{"message": "successfully created user"}))
}

// GenerateVerificationCodeHandler generates an verification code for an unverified user
func (app *Configs) GenerateVerificationCodeHandler(w http.ResponseWriter, r *http.Request) {
	var requestBody struct {
		Email string `json:"email"`
		validator.Validator
	}

	if err := helpers.ReadJSON(w, r, &requestBody); err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, helpers.ErrorResponse("unable to parse json body"))
		return
	}

	requestBody.CheckRequired(requestBody.Email, "email")
	requestBody.CheckValue(validator.IsEmail(requestBody.Email), "email", "valid email required")
	if !requestBody.Valid() {
		helpers.WriteJSON(w, http.StatusBadRequest, helpers.ErrorResponse(requestBody.Error()))
		return
	}

	user, err := app.DB.GetUser(r.Context(), requestBody.Email)
	if errors.Is(err, sql.ErrNoRows) {
		helpers.WriteJSON(w, http.StatusBadRequest, helpers.ErrorResponse("user does not exist"))
		return
	}

	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, helpers.ErrorResponse(err.Error()))
		return
	}

	if user.IsVerified() {
		helpers.WriteJSON(w, http.StatusBadRequest, helpers.ErrorResponse("user already verified"))
		return
	}

	verificationCode, err := app.Verifier.GenerateVerificationCode()
	if err != nil {
		helpers.WriteJSON(w, http.StatusInternalServerError, helpers.ErrorResponse(err.Error()))
		return
	}

	verification := models.Verification{
		Email:             requestBody.Email,
		VerificationType:  models.VerificationTypeAccount,
		VerificationCode:  verificationCode,
		ExpiresAt:         time.Now().Add(time.Hour * 24),
		AttemptsRemaining: app.Verifier.MaxRetries(),
	}

	if err := app.DB.InsertOrUpdateVerification(r.Context(), verification); err != nil {
		helpers.WriteJSON(w, http.StatusInternalServerError, helpers.ErrorResponse(err.Error()))
		return
	}

	var responseBody struct {
		VerificationCode string `json:"verification_code"`
	}

	responseBody.VerificationCode = verificationCode

	helpers.WriteJSON(w, http.StatusOK, helpers.SuccessResponse(responseBody))
}

// VerifyUserHandler is used to verify a user account given a valid verification code (provided by GenerateVerificationCodeHandler)
func (app *Configs) VerifyUserHandler(w http.ResponseWriter, r *http.Request) {
	var requestBody struct {
		Email            string `json:"email"`
		VerificationCode string `json:"verification_code"`
		validator.Validator
	}

	if err := helpers.ReadJSON(w, r, &requestBody); err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, helpers.ErrorResponse("unable to parse json body"))
		return
	}

	requestBody.CheckRequired(requestBody.Email, "email")
	requestBody.CheckValue(validator.IsEmail(requestBody.Email), "email", "valid email required")
	requestBody.CheckRequired(requestBody.VerificationCode, "verification_code")

	if !requestBody.Valid() {
		helpers.WriteJSON(w, http.StatusBadRequest, helpers.ErrorResponse(requestBody.Error()))
		return
	}

	user, err := app.DB.GetUser(r.Context(), requestBody.Email)
	if errors.Is(err, sql.ErrNoRows) {
		helpers.WriteJSON(w, http.StatusBadRequest, helpers.ErrorResponse("user does not exist"))
		return
	}

	if err != nil {
		helpers.WriteJSON(w, http.StatusInternalServerError, helpers.ErrorResponse(err.Error()))
		return
	}

	verification, err := app.DB.GetVerification(r.Context(), models.VerificationTypeAccount, requestBody.Email)
	if errors.Is(err, sql.ErrNoRows) {
		helpers.WriteJSON(w, http.StatusInternalServerError, helpers.ErrorResponse(fmt.Sprintf("no account verification data found for user %s", requestBody.Email)))
		return
	}

	if err != nil {
		helpers.WriteJSON(w, http.StatusInternalServerError, helpers.ErrorResponse(err.Error()))
		return
	}

	if verification.ExpiresAt.Before(time.Now()) || verification.AttemptsRemaining <= 0 {
		err := app.DB.DeleteVerification(r.Context(), requestBody.Email)
		if err != nil {
			app.Logger.Error(err.Error())
		}

		helpers.WriteJSON(w, http.StatusBadRequest, helpers.ErrorResponse("user verification code has expired"))
		return
	}

	if requestBody.VerificationCode != verification.VerificationCode {
		if verification.AttemptsRemaining >= 0 {
			if err := app.DB.InsertOrUpdateVerification(r.Context(), *verification); err != nil {
				helpers.WriteJSON(w, http.StatusInternalServerError, helpers.ErrorResponse(err.Error()))
				return
			}
		}

		helpers.WriteJSON(w, http.StatusBadRequest, helpers.ErrorResponse("invalid user verification code"))
		return
	}

	user.Status = models.UserStatusActive
	if err := app.DB.UpdateUser(r.Context(), *user); err != nil {
		helpers.WriteJSON(w, http.StatusInternalServerError, helpers.ErrorResponse(err.Error()))
		return
	}

	if err := app.DB.DeleteVerification(r.Context(), user.Email); err != nil {
		app.Logger.Error(err.Error())
	}

	helpers.WriteJSON(w, http.StatusOK, helpers.SuccessResponse(nil))
}

// TokenHandler verifies a user's email and password and returns a JWT (JSON Web Token) if valid credentials were provided
func (app *Configs) TokenHandler(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		validator.Validator
	}

	if err := helpers.ReadJSON(w, r, &body); err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, helpers.ErrorResponse("unable to parse json body"))
		return
	}

	body.CheckRequired(body.Email, "email")
	body.CheckRequired(body.Password, "password")
	body.CheckValue(validator.IsEmail(body.Email), "email", "valid email required")
	if !body.Valid() {
		helpers.WriteJSON(w, http.StatusBadRequest, helpers.ErrorResponse(body.Error()))
		return
	}

	user, err := app.DB.GetUser(r.Context(), body.Email)
	if errors.Is(err, sql.ErrNoRows) {
		helpers.WriteJSON(w, http.StatusBadRequest, helpers.ErrorResponse("invalid email or password"))
		return
	}

	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, helpers.ErrorResponse(err.Error()))
		return
	}

	if user.Status != models.UserStatusActive {
		helpers.WriteJSON(w, http.StatusBadRequest, helpers.ErrorResponse("user not active"))
		return
	}

	err = app.PasswordEncryptor.CompareHashAndPassword([]byte(user.Password), []byte(body.Password))
	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, helpers.ErrorResponse("invalid email or password"))
		return
	}

	tokenString, err := app.TokenUtils.GenerateToken(user.UserID, 24)
	if err != nil {
		helpers.WriteJSON(w, http.StatusInternalServerError, helpers.ErrorResponse("auth token generation failed"))
		return
	}

	helpers.WriteJSON(w, http.StatusOK, helpers.SuccessResponse(map[string]any{"token": tokenString}))
}

// DeleteUserHandler takes an email address and deletes the related user and verification data
func (app *Configs) DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email string `json:"email"`
		validator.Validator
	}

	if err := helpers.ReadJSON(w, r, &body); err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, helpers.ErrorResponse("unable to parse json body"))
		return
	}

	body.CheckRequired(body.Email, "email")
	body.CheckValue(validator.IsEmail(body.Email), "email", "valid email required")
	if !body.Valid() {
		helpers.WriteJSON(w, http.StatusBadRequest, helpers.ErrorResponse(body.Error()))
		return
	}

	recordsDeleted, err := app.DB.DeleteUser(r.Context(), body.Email)
	if err != nil {
		helpers.WriteJSON(w, http.StatusInternalServerError, helpers.ErrorResponse(err.Error()))
		return
	}

	if !recordsDeleted {
		helpers.WriteJSON(w, http.StatusOK, helpers.SuccessResponse(map[string]any{"message": "user not found"}))
		return
	}

	helpers.WriteJSON(w, http.StatusOK, helpers.SuccessResponse(nil))
}

func (app *Configs) ResetPasswordRequestHandler(w http.ResponseWriter, r *http.Request) {
	var requestBody struct {
		Email string `json:"email"`
		validator.Validator
	}

	if err := helpers.ReadJSON(w, r, &requestBody); err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, helpers.ErrorResponse("unable to parse json body"))
		return
	}

	requestBody.CheckRequired(requestBody.Email, "email")
	requestBody.CheckValue(validator.IsEmail(requestBody.Email), "email", "valid email required")
	if !requestBody.Valid() {
		helpers.WriteJSON(w, http.StatusBadRequest, helpers.ErrorResponse(requestBody.Error()))
		return
	}

	user, err := app.DB.GetUser(r.Context(), requestBody.Email)
	if errors.Is(err, sql.ErrNoRows) {
		helpers.WriteJSON(w, http.StatusBadRequest, helpers.ErrorResponse("user does not exist"))
		return
	}

	if user.Status != models.UserStatusActive {
		helpers.WriteJSON(w, http.StatusBadRequest, helpers.ErrorResponse("user is not active"))
		return
	}

	verificationCode, err := app.Verifier.GenerateVerificationCode()
	if err != nil {
		helpers.WriteJSON(w, http.StatusInternalServerError, helpers.ErrorResponse(err.Error()))
		return
	}

	verification := models.Verification{
		Email:             requestBody.Email,
		VerificationType:  models.VerificationTypeReset,
		VerificationCode:  verificationCode,
		ExpiresAt:         time.Now().Add(time.Hour * 24),
		AttemptsRemaining: app.Verifier.MaxRetries(),
	}

	user.Status = models.UserStatusVerifyResetPassword
	if err := app.DB.UpdateUser(r.Context(), *user); err != nil {
		helpers.WriteJSON(w, http.StatusInternalServerError, helpers.ErrorResponse(err.Error()))
		return
	}

	if err := app.DB.InsertOrUpdateVerification(r.Context(), verification); err != nil {
		helpers.WriteJSON(w, http.StatusInternalServerError, helpers.ErrorResponse(err.Error()))
		return
	}

	var responseBody struct {
		VerificationCode string `json:"verification_code"`
	}

	responseBody.VerificationCode = verificationCode

	helpers.WriteJSON(w, http.StatusOK, helpers.SuccessResponse(responseBody))
}

func (app *Configs) ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	var requestBody struct {
		Email            string `json:"email"`
		Password         string `json:"password"`
		VerificationCode string `json:"verification_code"`
		validator.Validator
	}

	if err := helpers.ReadJSON(w, r, &requestBody); err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, helpers.ErrorResponse("unable to parse json body"))
		return
	}

	requestBody.CheckRequired(requestBody.Email, "email")
	requestBody.CheckRequired(requestBody.Password, "password")
	requestBody.CheckValue(validator.IsEmail(requestBody.Email), "email", "valid email required")
	requestBody.CheckRequired(requestBody.VerificationCode, "verification_code")

	if !requestBody.Valid() {
		helpers.WriteJSON(w, http.StatusBadRequest, helpers.ErrorResponse(requestBody.Error()))
		return
	}

	user, err := app.DB.GetUser(r.Context(), requestBody.Email)
	if errors.Is(err, sql.ErrNoRows) {
		helpers.WriteJSON(w, http.StatusBadRequest, helpers.ErrorResponse("user does not exist"))
		return
	}

	if err != nil {
		helpers.WriteJSON(w, http.StatusInternalServerError, helpers.ErrorResponse(err.Error()))
		return
	}

	if user.Status != models.UserStatusVerifyResetPassword {
		helpers.WriteJSON(w, http.StatusBadRequest, helpers.ErrorResponse("password reset must first be requested"))
		return
	}

	verification, err := app.DB.GetVerification(r.Context(), models.VerificationTypeReset, requestBody.Email)
	if errors.Is(err, sql.ErrNoRows) {
		helpers.WriteJSON(w, http.StatusBadRequest, helpers.ErrorResponse(fmt.Sprintf("no password reset verification data found for user %s", requestBody.Email)))
		return
	}

	if err != nil {
		helpers.WriteJSON(w, http.StatusInternalServerError, helpers.ErrorResponse(err.Error()))
		return
	}

	if verification.ExpiresAt.Before(time.Now()) || verification.AttemptsRemaining <= 0 {
		err := app.DB.DeleteVerification(r.Context(), requestBody.Email)
		if err != nil {
			app.Logger.Error(err.Error())
		}

		helpers.WriteJSON(w, http.StatusBadRequest, helpers.ErrorResponse("password reset verification code has expired"))
		return
	}

	if requestBody.VerificationCode != verification.VerificationCode {
		if verification.AttemptsRemaining >= 0 {
			if err := app.DB.InsertOrUpdateVerification(r.Context(), *verification); err != nil {
				helpers.WriteJSON(w, http.StatusInternalServerError, helpers.ErrorResponse(err.Error()))
				return
			}
		}

		helpers.WriteJSON(w, http.StatusBadRequest, helpers.ErrorResponse("invalid password reset verification code"))
		return
	}

	hashedPasswordBytes, err := app.PasswordEncryptor.GenerateHashedPassword(requestBody.Password)
	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, helpers.ErrorResponse(err.Error()))
		return
	}

	user.Status = models.UserStatusActive
	user.Password = string(hashedPasswordBytes)
	if err := app.DB.UpdateUser(r.Context(), *user); err != nil {
		helpers.WriteJSON(w, http.StatusInternalServerError, helpers.ErrorResponse(err.Error()))
		return
	}

	if err := app.DB.DeleteVerification(r.Context(), user.Email); err != nil {
		app.Logger.Error(err.Error())
	}

	helpers.WriteJSON(w, http.StatusOK, helpers.SuccessResponse(nil))
}

func (app *Configs) UpdatePasswordHandler(w http.ResponseWriter, r *http.Request) {
	var requestBody struct {
		Email       string `json:"email"`
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
		validator.Validator
	}

	if err := helpers.ReadJSON(w, r, &requestBody); err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, helpers.ErrorResponse("unable to parse json body"))
		return
	}

	requestBody.CheckRequired(requestBody.Email, "email")
	requestBody.CheckRequired(requestBody.OldPassword, "old password")
	requestBody.CheckRequired(requestBody.NewPassword, "new password")
	requestBody.CheckValue(validator.IsEmail(requestBody.Email), "email", "valid email required")

	if !requestBody.Valid() {
		helpers.WriteJSON(w, http.StatusBadRequest, helpers.ErrorResponse(requestBody.Error()))
		return
	}

	user, err := app.DB.GetUser(r.Context(), requestBody.Email)
	if errors.Is(err, sql.ErrNoRows) {
		helpers.WriteJSON(w, http.StatusBadRequest, helpers.ErrorResponse("user does not exist"))
		return
	}

	if err != nil {
		helpers.WriteJSON(w, http.StatusInternalServerError, helpers.ErrorResponse(err.Error()))
		return
	}

	if user.Status != models.UserStatusActive {
		helpers.WriteJSON(w, http.StatusBadRequest, helpers.ErrorResponse("user not active"))
		return
	}
	err = app.PasswordEncryptor.CompareHashAndPassword([]byte(user.Password), []byte(requestBody.OldPassword))
	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, helpers.ErrorResponse("old password is invalid"))
		return
	}

	hashedPasswordBytes, err := app.PasswordEncryptor.GenerateHashedPassword(requestBody.NewPassword)
	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, helpers.ErrorResponse(err.Error()))
		return
	}

	user.Password = string(hashedPasswordBytes)
	if err := app.DB.UpdateUser(r.Context(), *user); err != nil {
		helpers.WriteJSON(w, http.StatusInternalServerError, helpers.ErrorResponse(err.Error()))
		return
	}

	helpers.WriteJSON(w, http.StatusOK, helpers.SuccessResponse(map[string]any{"message": "successfully updated password"}))
}

func (app *Configs) UserRoleHandler(w http.ResponseWriter, r *http.Request) {
	var requestBody struct {
		Email string `json:"email"`
		validator.Validator
	}

	if err := helpers.ReadJSON(w, r, &requestBody); err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, helpers.ErrorResponse("unable to parse json body"))
		return
	}

	requestBody.CheckRequired(requestBody.Email, "email")
	requestBody.CheckValue(validator.IsEmail(requestBody.Email), "email", "valid email required")

	if !requestBody.Valid() {
		helpers.WriteJSON(w, http.StatusBadRequest, helpers.ErrorResponse(requestBody.Error()))
		return
	}

	user, err := app.DB.GetUser(r.Context(), requestBody.Email)
	if errors.Is(err, sql.ErrNoRows) {
		helpers.WriteJSON(w, http.StatusBadRequest, helpers.ErrorResponse("user does not exist"))
		return
	}

	if err != nil {
		helpers.WriteJSON(w, http.StatusInternalServerError, helpers.ErrorResponse(err.Error()))
		return
	}

	var respBody struct {
		Role string `json:"role"`
	}

	respBody.Role = user.Role

	helpers.WriteJSON(w, http.StatusOK, helpers.SuccessResponse(respBody))
}
