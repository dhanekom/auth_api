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
			req, _ := http.NewRequest(http.MethodPost, versionUrl("/api/auth/register"), nil)
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
			req, _ := http.NewRequest(http.MethodDelete, versionUrl("/api/auth/user"), nil)
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
		{desc: "missing parameters", reqBody: `{}`, status: http.StatusBadRequest, want: `{"status":"error","message":"email: required, password: required"}`},
		{desc: "invalid email", reqBody: `{"email": "invalidemail", "password": "1234"}`, status: http.StatusBadRequest, want: `{"status":"error","message":"email: valid email required"}`},
		{desc: "user already exists", reqBody: `{"email": "unverified@gmail.com", "password": "1234"}`, status: http.StatusBadRequest, want: `{"status":"error","message":"user already exists"}`},
		{desc: "success", reqBody: `{"email": "notexist@gmail.com", "password": "1234"}`, status: http.StatusOK, want: `{"status":"success","data":{"message":"successfully created user"}}`},
	}
	ctx := context.Background()
	app := setupApp(t, ctx)

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodPost, versionUrl("/api/auth/register"), strings.NewReader(test.reqBody))
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
		{desc: "invalid email", reqBody: `{"email": "invalidemail", "password": "1234"}`, status: http.StatusBadRequest, want: `{"status":"error","message":"email: valid email required"}`},
		{desc: "user does not exist", reqBody: `{"email": "notexist@gmail.com"}`, status: http.StatusBadRequest, want: `{"status":"error","message":"user does not exist"}`},
		{desc: "already verified", reqBody: `{"email": "verified@gmail.com"}`, status: http.StatusBadRequest, want: `{"status":"error","message":"user already verified"}`},
		{desc: "success", reqBody: `{"email": "unverified@gmail.com"}`, status: http.StatusOK, want: `{"status":"success","data":{"verification_code":"ABCDEF"}}`},
	}

	ctx := context.Background()
	app := setupApp(t, ctx)

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, versionUrl("/api/auth/verify"), strings.NewReader(test.reqBody))
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
		{desc: "no verification data", reqBody: `{"email": "noverification@gmail.com", "verification_code": "ABCDEF"}`, status: http.StatusInternalServerError, want: `{"status":"error","message":"no verification data found for user noverification@gmail.com"}`},
		{desc: "verification code has expired", reqBody: `{"email": "expiredverification@gmail.com", "verification_code": "ABCDEF"}`, status: http.StatusBadRequest, want: `{"status":"error","message":"verification code has expired"}`},
		{desc: "too many attempts", reqBody: `{"email": "toomanyattempts@gmail.com", "verification_code": "ABCDEF"}`, status: http.StatusBadRequest, want: `{"status":"error","message":"verification code has expired"}`},
		{desc: "invalid verification code", reqBody: `{"email": "unverified@gmail.com", "verification_code": "INVALID"}`, status: http.StatusBadRequest, want: `{"status":"error","message":"invalid verification code"}`},
		{desc: "success", reqBody: `{"email": "unverified@gmail.com", "verification_code": "ABCDEF"}`, status: http.StatusOK, want: `{"status":"success"}`},
	}

	ctx := context.Background()
	app := setupApp(t, ctx)

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodPost, versionUrl("/api/auth/verify"), strings.NewReader(test.reqBody))
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
		{desc: "user does not exist", reqBody: `{"email": "notexit@gmail.com", "password": "1234"}`, status: http.StatusUnauthorized, want: `{"status":"error","message":"invalid email or password"}`},
		{desc: "user not verified", reqBody: `{"email": "unverified@gmail.com", "password": "1234"}`, status: http.StatusUnauthorized, want: `{"status":"error","message":"user not verified"}`},
		{desc: "invalid password", reqBody: `{"email": "invalidpassword@gmail.com", "password": "invalid"}`, status: http.StatusUnauthorized, want: `{"status":"error","message":"invalid email or password"}`},
		{desc: "auth token generation failed", reqBody: `{"email": "authcodefailed@gmail.com", "password": "validpass"}`, status: http.StatusBadRequest, want: `{"status":"error","message":"auth token generation failed"}`},
		{desc: "success", reqBody: `{"email": "verified@gmail.com", "password": "1234"}`, status: http.StatusOK, want: fmt.Sprintf(`{"status":"success","data":{"token":"%s"}}`, TestToken)},
	}

	ctx := context.Background()
	app := setupApp(t, ctx)

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodPost, versionUrl("/api/auth/token"), strings.NewReader(test.reqBody))
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
			req, _ := http.NewRequest(http.MethodDelete, versionUrl("/api/auth/user"), strings.NewReader(test.reqBody))
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
	req, _ := http.NewRequest(http.MethodGet, versionUrl("/api/auth/healthz"), nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", userAuthToken))
	w := httptest.NewRecorder()
	app.server.Handler.ServeHTTP(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
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
