package main

import (
	"auth_api/internal/verify"
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	apiVersion     = "v1"
	userAuthToken  = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkRpZmYiLCJpYXQiOjE1MTYyMzkwMjJ9.6Xq-5W9lU5IVp0iCnSiBIuvoBaxfi7V4vbxRzK-H0YM"
	adminAuthToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkRpZmYiLCJpYXQiOjE1MTYyMzkwMjJ9.8SgpGNj3y4HI8884-yuMszgRghYS0j6i7KP1OOLpuMg"
)

func versionUrl(aURL string) string {
	return fmt.Sprintf("/%s%s", apiVersion, aURL)
}

func TestAuthMiddelwareBlockAccess(t *testing.T) {
	tests := []struct {
		desc       string
		authHeader string
		statusCode int
		want       string
	}{
		{desc: "no authorization header", authHeader: "", statusCode: http.StatusUnauthorized, want: `{"status":"error","message":"authorization failed"}`},
		{desc: "invalid authorization header", authHeader: "bearer asdf aa", statusCode: http.StatusUnauthorized, want: `{"status":"error","message":"authorization failed"}`},
		{desc: "invalid bearer token", authHeader: "Bearer invalidtoken", statusCode: http.StatusUnauthorized, want: `{"status":"error","message":"token verification failed"}`},
	}

	ctx := context.Background()
	app := setupApp(t, ctx)
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodPost, versionUrl("/auth/register"), nil)
			if test.authHeader != "" {
				req.Header.Set("Authorization", test.authHeader)
			}
			w := httptest.NewRecorder()
			app.server.Handler.ServeHTTP(w, req)

			resp := w.Result()
			json, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Errorf("didn't expect error but got %s", err)
			}

			assert.Equal(t, test.statusCode, resp.StatusCode)
			assert.Equal(t, test.want, string(json))
		})
	}
}

func TestAdminMiddelwareBlockAccess(t *testing.T) {
	tests := []struct {
		desc       string
		authHeader string
		statusCode int
		want       string
	}{
		{desc: "no authorization header", authHeader: "", statusCode: http.StatusUnauthorized, want: `{"status":"error","message":"authorization failed"}`},
		{desc: "invalid bearer token", authHeader: fmt.Sprintf("Bearer %s", userAuthToken), statusCode: http.StatusForbidden, want: `{"status":"error","message":"admin access rights required"}`},
	}

	ctx := context.Background()
	app := setupApp(t, ctx)
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodDelete, versionUrl("/admin/auth/user"), nil)
			if test.authHeader != "" {
				req.Header.Set("Authorization", test.authHeader)
			}
			w := httptest.NewRecorder()
			app.server.Handler.ServeHTTP(w, req)

			resp := w.Result()
			json, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Errorf("didn't expect error but got %s", err)
			}

			assert.Equal(t, test.statusCode, resp.StatusCode)
			assert.Equal(t, test.want, string(json))
		})
	}
}

func TestRegisterHandler(t *testing.T) {
	tests := []struct {
		desc    string
		reqBody string
		status  int
		want    string
	}{
		{desc: "invalid request json body", reqBody: ``, status: http.StatusBadRequest, want: `{"status":"error","message":"unable to parse json body"}`},
		{desc: "missing parameters", reqBody: `{}`, status: http.StatusBadRequest, want: `{"status":"error","message":"email: required, password: required, role: required"}`},
		{desc: "invalid email", reqBody: `{"email": "invalidemail", "password": "1234", "role": "USER"}`, status: http.StatusBadRequest, want: `{"status":"error","message":"email: valid email required"}`},
		{desc: "user already exists", reqBody: `{"email": "unverified@gmail.com", "password": "1234", "role": "USER"}`, status: http.StatusBadRequest, want: `{"status":"error","message":"user already exists"}`},
		{desc: "success", reqBody: `{"email": "notexist@gmail.com", "password": "1234", "role": "USER"}`, status: http.StatusOK, want: `{"status":"success","data":{"message":"successfully created user"}}`},
	}
	ctx := context.Background()
	app := setupApp(t, ctx)

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodPost, versionUrl("/auth/register"), strings.NewReader(test.reqBody))
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", userAuthToken))
			w := httptest.NewRecorder()
			app.server.Handler.ServeHTTP(w, req)

			resp := w.Result()
			json, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Errorf("didn't expect error but got %s", err)
			}

			assert.Equal(t, test.status, resp.StatusCode)
			assert.Equal(t, test.want, string(json))
		})
	}
}

func TestGenerateVerificationCodeHandler(t *testing.T) {
	tests := []struct {
		desc    string
		reqBody string
		status  int
		want    string
	}{
		{desc: "invalid request json body", reqBody: ``, status: http.StatusBadRequest, want: `{"status":"error","message":"unable to parse json body"}`},
		{desc: "missing parameters", reqBody: `{}`, status: http.StatusBadRequest, want: `{"status":"error","message":"email: required"}`},
		{desc: "invalid email", reqBody: `{"email": "invalidemail", "password": "1234", "role": "USER"}`, status: http.StatusBadRequest, want: `{"status":"error","message":"email: valid email required"}`},
		{desc: "user does not exist", reqBody: `{"email": "notexist@gmail.com"}`, status: http.StatusBadRequest, want: `{"status":"error","message":"user does not exist"}`},
		{desc: "already verified", reqBody: `{"email": "verified@gmail.com"}`, status: http.StatusBadRequest, want: `{"status":"error","message":"user already verified"}`},
		{desc: "success", reqBody: `{"email": "unverified@gmail.com"}`, status: http.StatusOK, want: `{"status":"success","data":{"verification_code":"ABCDEF"}}`},
	}

	ctx := context.Background()
	app := setupApp(t, ctx)

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, versionUrl("/auth/verifyuser"), strings.NewReader(test.reqBody))
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", userAuthToken))
			w := httptest.NewRecorder()
			app.server.Handler.ServeHTTP(w, req)

			resp := w.Result()
			json, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Errorf("didn't expect error but got %s", err)
			}

			assert.Equal(t, test.status, resp.StatusCode)
			assert.Equal(t, test.want, string(json))
		})
	}
}

func TestVerifyUserHandler(t *testing.T) {
	tests := []struct {
		desc    string
		reqBody string
		status  int
		want    string
	}{
		{desc: "invalid request json body", reqBody: ``, status: http.StatusBadRequest, want: `{"status":"error","message":"unable to parse json body"}`},
		{desc: "missing parameters", reqBody: `{}`, status: http.StatusBadRequest, want: `{"status":"error","message":"email: required, verification_code: required"}`},
		{desc: "invalid email", reqBody: `{"email": "invalidemail", "verification_code": "ABCDEF"}`, status: http.StatusBadRequest, want: `{"status":"error","message":"email: valid email required"}`},
		{desc: "user does not exist", reqBody: `{"email": "notexist@gmail.com", "verification_code": "ABCDEF"}`, status: http.StatusBadRequest, want: `{"status":"error","message":"user does not exist"}`},
		{desc: "no account verification data", reqBody: `{"email": "noverification@gmail.com", "verification_code": "ABCDEF"}`, status: http.StatusInternalServerError, want: `{"status":"error","message":"no account verification data found for user noverification@gmail.com"}`},
		{desc: "user verification code has expired", reqBody: `{"email": "expiredverification@gmail.com", "verification_code": "ABCDEF"}`, status: http.StatusBadRequest, want: `{"status":"error","message":"user verification code has expired"}`},
		{desc: "too many attempts", reqBody: `{"email": "toomanyattempts@gmail.com", "verification_code": "ABCDEF"}`, status: http.StatusBadRequest, want: `{"status":"error","message":"user verification code has expired"}`},
		{desc: "invalid user verification code", reqBody: `{"email": "unverified@gmail.com", "verification_code": "INVALID"}`, status: http.StatusBadRequest, want: `{"status":"error","message":"invalid user verification code"}`},
		{desc: "success", reqBody: `{"email": "unverified@gmail.com", "verification_code": "ABCDEF"}`, status: http.StatusOK, want: `{"status":"success"}`},
	}

	ctx := context.Background()
	app := setupApp(t, ctx)

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodPost, versionUrl("/auth/verifyuser"), strings.NewReader(test.reqBody))
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", userAuthToken))
			w := httptest.NewRecorder()
			app.server.Handler.ServeHTTP(w, req)

			resp := w.Result()
			json, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Errorf("didn't expect error but got %s", err)
			}

			assert.Equal(t, test.status, resp.StatusCode)
			assert.Equal(t, test.want, string(json))
		})
	}
}

func TestTokenHandler(t *testing.T) {
	tests := []struct {
		desc    string
		reqBody string
		status  int
		want    string
	}{
		{desc: "invalid request json body", reqBody: ``, status: http.StatusBadRequest, want: `{"status":"error","message":"unable to parse json body"}`},
		{desc: "missing parameters", reqBody: `{}`, status: http.StatusBadRequest, want: `{"status":"error","message":"email: required, password: required"}`},
		{desc: "invalid email", reqBody: `{"email": "invalidemail", "password": "1234"}`, status: http.StatusBadRequest, want: `{"status":"error","message":"email: valid email required"}`},
		{desc: "user does not exist", reqBody: `{"email": "notexit@gmail.com", "password": "1234"}`, status: http.StatusBadRequest, want: `{"status":"error","message":"invalid email or password"}`},
		{desc: "user not verified", reqBody: `{"email": "unverified@gmail.com", "password": "1234"}`, status: http.StatusBadRequest, want: `{"status":"error","message":"user not active"}`},
		{desc: "invalid password", reqBody: `{"email": "invalidpassword@gmail.com", "password": "invalid"}`, status: http.StatusBadRequest, want: `{"status":"error","message":"invalid email or password"}`},
		{desc: "auth token generation failed", reqBody: `{"email": "authcodefailed@gmail.com", "password": "validpass"}`, status: http.StatusInternalServerError, want: `{"status":"error","message":"auth token generation failed"}`},
		{desc: "success", reqBody: `{"email": "verified@gmail.com", "password": "1234"}`, status: http.StatusOK, want: fmt.Sprintf(`{"status":"success","data":{"token":"%s"}}`, TestToken)},
	}

	ctx := context.Background()
	app := setupApp(t, ctx)

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodPost, versionUrl("/auth/token"), strings.NewReader(test.reqBody))
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", userAuthToken))
			w := httptest.NewRecorder()
			app.server.Handler.ServeHTTP(w, req)

			resp := w.Result()
			json, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Errorf("didn't expect error but got %s", err)
			}

			assert.Equal(t, test.status, resp.StatusCode)
			assert.Equal(t, test.want, string(json))
		})
	}
}

func TestDeleteUserHandler(t *testing.T) {
	tests := []struct {
		desc    string
		reqBody string
		status  int
		want    string
	}{
		{desc: "invalid request json body", reqBody: ``, status: http.StatusBadRequest, want: `{"status":"error","message":"unable to parse json body"}`},
		{desc: "missing parameters", reqBody: `{}`, status: http.StatusBadRequest, want: `{"status":"error","message":"email: required"}`},
		{desc: "invalid email", reqBody: `{"email": "invalidemail", "password": "1234"}`, status: http.StatusBadRequest, want: `{"status":"error","message":"email: valid email required"}`},
		{desc: "user not found", reqBody: `{"email": "notfound@gmail.com", "password": "1234"}`, status: http.StatusOK, want: `{"status":"success","data":{"message":"user not found"}}`},
		{desc: "success", reqBody: `{"email": "verified@gmail.com", "password": "1234"}`, status: http.StatusOK, want: `{"status":"success"}`},
	}

	ctx := context.Background()
	app := setupApp(t, ctx)

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodDelete, versionUrl("/admin/auth/user"), strings.NewReader(test.reqBody))
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", adminAuthToken))
			w := httptest.NewRecorder()
			app.server.Handler.ServeHTTP(w, req)

			resp := w.Result()
			json, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Errorf("didn't expect error but got %s", err)
			}

			assert.Equal(t, test.status, resp.StatusCode)
			assert.Equal(t, test.want, string(json))
		})
	}
}

func TestHealthzHandler(t *testing.T) {
	ctx := context.Background()
	app := setupApp(t, ctx)
	req, _ := http.NewRequest(http.MethodGet, versionUrl("/auth/healthz"), nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", userAuthToken))
	w := httptest.NewRecorder()
	app.server.Handler.ServeHTTP(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestResetPasswordRequestHandler(t *testing.T) {
	tests := []struct {
		desc    string
		reqBody string
		status  int
		want    string
	}{
		{desc: "invalid request json body", reqBody: ``, status: http.StatusBadRequest, want: `{"status":"error","message":"unable to parse json body"}`},
		{desc: "missing parameters", reqBody: `{}`, status: http.StatusBadRequest, want: `{"status":"error","message":"email: required"}`},
		{desc: "invalid email", reqBody: `{"email": "invalidemail"}`, status: http.StatusBadRequest, want: `{"status":"error","message":"email: valid email required"}`},
		{desc: "user does not exist", reqBody: `{"email": "notexist@gmail.com"}`, status: http.StatusBadRequest, want: `{"status":"error","message":"user does not exist"}`},
		{desc: "user not active", reqBody: `{"email": "unverified@gmail.com"}`, status: http.StatusBadRequest, want: `{"status":"error","message":"user is not active"}`},
		{desc: "success", reqBody: `{"email": "resetpasswordrequest@gmail.com"}`, status: http.StatusOK, want: `{"status":"success","data":{"verification_code":"ABCDEF"}}`},
	}

	ctx := context.Background()
	app := setupApp(t, ctx)

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodPost, versionUrl("/auth/resetpassword"), strings.NewReader(test.reqBody))
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", userAuthToken))
			w := httptest.NewRecorder()
			app.server.Handler.ServeHTTP(w, req)

			resp := w.Result()
			json, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Errorf("didn't expect error but got %s", err)
			}

			assert.Equal(t, test.status, resp.StatusCode)
			assert.Equal(t, test.want, string(json))
		})
	}
}

func TestResetPasswordHandler(t *testing.T) {
	tests := []struct {
		desc    string
		reqBody string
		status  int
		want    string
	}{
		{desc: "invalid request json body", reqBody: ``, status: http.StatusBadRequest, want: `{"status":"error","message":"unable to parse json body"}`},
		{desc: "missing parameters", reqBody: `{}`, status: http.StatusBadRequest, want: `{"status":"error","message":"email: required, password: required, verification_code: required"}`},
		{desc: "invalid email", reqBody: `{"email": "invalidemail", "password": "invalidpass", "verification_code": "ABCDEF"}`, status: http.StatusBadRequest, want: `{"status":"error","message":"email: valid email required"}`},
		{desc: "user does not exist", reqBody: `{"email": "notexist@gmail.com", "password": "invalidpass", "verification_code": "ABCDEF"}`, status: http.StatusBadRequest, want: `{"status":"error","message":"user does not exist"}`},
		{desc: "user status is not verify_reset", reqBody: `{"email": "notverifyresetstatus@gmail.com", "password": "invalidpass", "verification_code": "ABCDEF"}`, status: http.StatusBadRequest, want: `{"status":"error","message":"password reset must first be requested"}`},
		{desc: "no password reset verification data", reqBody: `{"email": "noresetverification@gmail.com", "password": "invalidpass", "verification_code": "ABCDEF"}`, status: http.StatusBadRequest, want: `{"status":"error","message":"no password reset verification data found for user noresetverification@gmail.com"}`},
		{desc: "success", reqBody: `{"email": "resetpassword@gmail.com", "password": "validpass", "verification_code": "ABCDEF"}`, status: http.StatusOK, want: `{"status":"success"}`},
	}

	ctx := context.Background()
	app := setupApp(t, ctx)

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodPut, versionUrl("/auth/resetpassword"), strings.NewReader(test.reqBody))
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", userAuthToken))
			w := httptest.NewRecorder()
			app.server.Handler.ServeHTTP(w, req)

			resp := w.Result()
			json, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Errorf("didn't expect error but got %s", err)
			}

			assert.Equal(t, test.status, resp.StatusCode)
			assert.Equal(t, test.want, string(json))
		})
	}
}

func TestUpdatePasswordHandler(t *testing.T) {
	tests := []struct {
		desc    string
		reqBody string
		status  int
		want    string
	}{
		{desc: "invalid request json body", reqBody: ``, status: http.StatusBadRequest, want: `{"status":"error","message":"unable to parse json body"}`},
		{desc: "missing parameters", reqBody: `{}`, status: http.StatusBadRequest, want: `{"status":"error","message":"email: required, old password: required, new password: required"}`},
		{desc: "invalid email", reqBody: `{"email": "invalidemail", "old_password": "1234", "new_password": "2345"}`, status: http.StatusBadRequest, want: `{"status":"error","message":"email: valid email required"}`},
		{desc: "user does not exist", reqBody: `{"email": "notexists@gmail.com", "old_password": "1234", "new_password": "2345"}`, status: http.StatusBadRequest, want: `{"status":"error","message":"user does not exist"}`},
		{desc: "user not active", reqBody: `{"email": "unverified@gmail.com", "old_password": "1234", "new_password": "2345"}`, status: http.StatusBadRequest, want: `{"status":"error","message":"user not active"}`},
		{desc: "success", reqBody: `{"email": "verified@gmail.com", "old_password": "9999", "new_password": "2345"}`, status: http.StatusOK, want: `{"status":"success","data":{"message":"successfully updated password"}}`},
	}
	ctx := context.Background()
	app := setupApp(t, ctx)

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodPost, versionUrl("/auth/updatepassword"), strings.NewReader(test.reqBody))
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", userAuthToken))
			w := httptest.NewRecorder()
			app.server.Handler.ServeHTTP(w, req)

			resp := w.Result()
			json, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Errorf("didn't expect error but got %s", err)
			}

			assert.Equal(t, test.status, resp.StatusCode)
			assert.Equal(t, test.want, string(json))
		})
	}
}

func TestUserRoleHandler(t *testing.T) {
	tests := []struct {
		desc    string
		reqBody string
		status  int
		want    string
	}{
		{desc: "invalid request json body", reqBody: ``, status: http.StatusBadRequest, want: `{"status":"error","message":"unable to parse json body"}`},
		{desc: "missing parameters", reqBody: `{}`, status: http.StatusBadRequest, want: `{"status":"error","message":"email: required"}`},
		{desc: "invalid email", reqBody: `{"email": "invalidemail"}`, status: http.StatusBadRequest, want: `{"status":"error","message":"email: valid email required"}`},
		{desc: "user does not exist", reqBody: `{"email": "notexists@gmail.com"}`, status: http.StatusBadRequest, want: `{"status":"error","message":"user does not exist"}`},
		{desc: "success", reqBody: `{"email": "verified@gmail.com"}`, status: http.StatusOK, want: `{"status":"success","data":{"Role":"USER"}}`},
	}
	ctx := context.Background()
	app := setupApp(t, ctx)

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, versionUrl("/auth/role"), strings.NewReader(test.reqBody))
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", userAuthToken))
			w := httptest.NewRecorder()
			app.server.Handler.ServeHTTP(w, req)

			resp := w.Result()
			json, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Errorf("didn't expect error but got %s", err)
			}

			assert.Equal(t, test.status, resp.StatusCode)
			assert.Equal(t, test.want, string(json))
		})
	}
}

func GetTestEnv(key string) string {
	switch key {
	case "AUTH_HOST_ADDR":
		return "localhost"
	case "AUTH_HOST_PORT":
		return "80"
	case "AUTH_DB_CONNECTION_STRING":
		return ""
	case "AUTH_JWT_SECRET":
		return "e63be6cb5ff205cc08b5fb1f8d2d67e2a6b4e8a21432b6236260c586526271657c1cca677f95e51dfd64f8c4c62383d45abc7af77025eb55dab03abc4eec04b27732fb0a7eeeb4db8b05bf0278d6305eb5a247957071850da50235d09af9fab3e2e32bdd5e67a67bb461fa11bd3ed081fd34d038841547bbfa079631fbda92aa73b569b3cb1417ec5fbdc01b82abb46ffa73cee613abcb5a1c8b4e441fe01ca46007d1b5ecc2d48ed573049db76998b51d27b23512b2f3199da039b7859395120bef26d9f56f6cfb6bd93fbbcfa732ab2651c76e22d3e7987ed31a5f754e3e6f2068107c61b707f557d00bc5431abaa4f19ed276e0a58b1821b164cffe267d4f"
	case "AUTH_USER_TOKEN_SECRET":
		return "usertokensecret"
	case "AUTH_ADMIN_TOKEN_SECRET":
		return "admintokensecret"
	default:
		return ""
	}
}

func setupApp(t *testing.T, ctx context.Context) *App {
	t.Helper()

	pgContainer, err := postgres.Run(
		ctx,
		"postgres:15.3-alpine",
		postgres.WithInitScripts(filepath.Join("..", "..", "migrations", "00000000_000000_init.up.sql")),
		postgres.WithInitScripts(filepath.Join("..", "..", "testing", "testdata", "init-db.sql")),
		postgres.WithDatabase("auth_db"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate pgContainer: %s", err)
		}
	})

	dbConnStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get ConnectionString: %s", err)
	}

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)
	app, err := NewServer(writer, GetTestEnv, dbConnStr, &MockUserVerifier{maxRetries: 3, verificationCode: "ABCDEF"}, &MockPasswordEncryptor{}, &MockTokenGenerator{})
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	return app
}

type MockUserVerifier struct {
	maxRetries       int
	verificationCode string
}

func (v *MockUserVerifier) Setup(codeLength, maximumRetries int) {
	if maximumRetries == 0 {
		maximumRetries = verify.DefaultMaxRetries
	}

	v.maxRetries = maximumRetries
}

func (v MockUserVerifier) MaxRetries() int {
	return v.maxRetries
}

func (v MockUserVerifier) GenerateVerificationCode() (string, error) {
	return v.verificationCode, nil
}

type MockPasswordEncryptor struct {
}

func (e MockPasswordEncryptor) GenerateHashedPassword(password string) ([]byte, error) {
	return []byte("validpass"), nil
}

func (e MockPasswordEncryptor) CompareHashAndPassword(hashedPassword, password []byte) error {
	if string(hashedPassword) != "validpass" {
		return errors.New("passwords don't match")
	}

	return nil
}

const TestToken = "dub8CuDY6VA6TdoHM9ViSpcSVS7R1I"

type MockTokenGenerator struct {
}

func (t *MockTokenGenerator) Setup(secret string) {
	//
}

func (t *MockTokenGenerator) GenerateToken(userID string, hours int) (string, error) {
	if userID == "0460d39a-9c81-48bd-86ed-7154f44ac617" {
		return "", errors.New("GenerateToken - unable to generate token")
	}
	return TestToken, nil
}
