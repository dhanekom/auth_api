package main

import (
	"auth_api/internal/verify"
	"bufio"
	"bytes"
	"context"
	"errors"
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
		{desc: "success", reqBody: `{"email": "test@gmail.com", "password": "1234"}`, status: http.StatusOK, want: `{"status":"success","data":{"message":"successfully created user"}}`},
	}
	ctx := context.Background()

	app := setupApp(t, ctx)

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {

			req, _ := http.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(test.reqBody))
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

// func TestGenerateVerificationCodeHandler(t *testing.T) {
// 	tests := []struct {
// 		desc    string
// 		reqBody string
// 		status  int
// 		want    string
// 		db      database.MockDBRepo
// 	}{
// 		{desc: "success", reqBody: `{"email": "test@gmail.com"}`, status: http.StatusOK, want: `{"status":"success","data":{"verification_code":"ABCDEF"}}`, db: database.MockDBRepo{
// 			TestUser: models.User{
// 				UserID:     1,
// 				Email:      "test@gmail.com",
// 				Password:   "1234",
// 				IsVerified: false,
// 			},
// 		}},
// 		{desc: "already verified", reqBody: `{"email": "test@gmail.com"}`, status: http.StatusBadRequest, want: `{"status":"error","message":"user already verified"}`, db: database.MockDBRepo{
// 			TestUser: models.User{
// 				UserID:     1,
// 				Email:      "test@gmail.com",
// 				Password:   "1234",
// 				IsVerified: true,
// 			},
// 		}},
// 		{desc: "invalid request json body", reqBody: ``, status: http.StatusBadRequest, want: `{"status":"error","message":"unable to parse json body"}`, db: database.MockDBRepo{}},
// 		{desc: "missing parameters", reqBody: `{}`, status: http.StatusBadRequest, want: `{"status":"error","message":"email: required"}`, db: database.MockDBRepo{}},
// 		{desc: "invalid email", reqBody: `{"email": "invalidemail", "password": "1234"}`, status: http.StatusBadRequest, want: `{"status":"error","message":"email: valid email required"}`, db: database.MockDBRepo{}},
// 		{desc: "user does not exist", reqBody: `{"email": "test@gmail.com"}`, status: http.StatusBadRequest, want: `{"status":"error","message":"user does not exist"}`, db: database.MockDBRepo{
// 			TestUser: models.User{
// 				UserID: 0,
// 			},
// 		}},
// 		{desc: "unable to insert verification record", reqBody: `{"email": "fail@gmail.com"}`, status: http.StatusInternalServerError, want: `{"status":"error","message":"InsertOrUpdateVerification failed"}`, db: database.MockDBRepo{
// 			TestUser: models.User{
// 				UserID: 1,
// 			},
// 		}},
// 	}

// 	for _, test := range tests {
// 		t.Run(test.desc, func(t *testing.T) {
// 			app := setupApp(&test.db)
// 			path := "/api/auth/verify"
// 			router := gin.New()
// 			router.GET(path, app.GenerateVerificationCodeHandler)

// 			body := strings.NewReader(test.reqBody)
// 			req := httptest.NewRequest(http.MethodGet, path, body)
// 			w := httptest.NewRecorder()
// 			router.ServeHTTP(w, req)

// 			resp := w.Result()
// 			json, err := io.ReadAll(resp.Body)
// 			if err != nil {
// 				t.Errorf("didn't expect error but got %s", err)
// 			}

// 			assert.Equal(t, test.status, resp.StatusCode)
// 			assert.Equal(t, test.want, string(json))
// 		})
// 	}
// }

// func TestVerifyUserHandler(t *testing.T) {
// 	tests := []struct {
// 		desc    string
// 		reqBody string
// 		status  int
// 		want    string
// 		db      database.MockDBRepo
// 	}{
// 		{desc: "success", reqBody: `{"email": "test@gmail.com", "verification_code": "ABCDEF"}`, status: http.StatusOK, want: `{"status":"success"}`, db: database.MockDBRepo{
// 			TestUser:         models.User{UserID: 1, Email: "test@gmail.com"},
// 			TestVerification: models.Verification{Email: "test@gmail.com", VerificationCode: "ABCDEF", ExpiresAt: time.Now().Add(time.Hour), AttemptsRemaining: 3},
// 		}},
// 		{desc: "invalid request json body", reqBody: ``, status: http.StatusBadRequest, want: `{"status":"error","message":"unable to parse json body"}`, db: database.MockDBRepo{}},
// 		{desc: "missing parameters", reqBody: `{}`, status: http.StatusBadRequest, want: `{"status":"error","message":"email: required, verification_code: required"}`, db: database.MockDBRepo{}},
// 		{desc: "invalid email", reqBody: `{"email": "invalidemail", "verification_code": "ABCDEF"}`, status: http.StatusBadRequest, want: `{"status":"error","message":"email: valid email required"}`, db: database.MockDBRepo{}},
// 		{desc: "user does not exist", reqBody: `{"email": "test@gmail.com", "verification_code": "ABCDEF"}`, status: http.StatusBadRequest, want: `{"status":"error","message":"user does not exist"}`, db: database.MockDBRepo{
// 			TestUser: models.User{UserID: 0, Email: "test@gmail.com"},
// 		}},
// 		{desc: "user does not exist", reqBody: `{"email": "fail@gmail.com", "verification_code": "ABCDEF"}`, status: http.StatusInternalServerError, want: `{"status":"error","message":"no verification data found for user fail@gmail.com"}`, db: database.MockDBRepo{
// 			TestUser:         models.User{UserID: 1, Email: "fail@gmail.com"},
// 			TestVerification: models.Verification{Email: "fail@gmail.com"},
// 		}},
// 		{desc: "verification code has expired", reqBody: `{"email": "test@gmail.com", "verification_code": "ABCDEF"}`, status: http.StatusBadRequest, want: `{"status":"error","message":"verification code has expired"}`, db: database.MockDBRepo{
// 			TestUser:         models.User{UserID: 1, Email: "test@gmail.com"},
// 			TestVerification: models.Verification{Email: "test@gmail.com", VerificationCode: "ABCDEF", ExpiresAt: time.Now().Add(time.Hour * -1), AttemptsRemaining: 3},
// 		}},
// 		{desc: "too many attempts", reqBody: `{"email": "test@gmail.com", "verification_code": "ABCDEF"}`, status: http.StatusBadRequest, want: `{"status":"error","message":"verification code has expired"}`, db: database.MockDBRepo{
// 			TestUser:         models.User{UserID: 1, Email: "test@gmail.com"},
// 			TestVerification: models.Verification{Email: "test@gmail.com", VerificationCode: "ABCDEF", ExpiresAt: time.Now().Add(time.Hour), AttemptsRemaining: 0},
// 		}},
// 		{desc: "invalid verification code", reqBody: `{"email": "test@gmail.com", "verification_code": "INVALID"}`, status: http.StatusBadRequest, want: `{"status":"error","message":"invalid verification code"}`, db: database.MockDBRepo{
// 			TestUser:         models.User{UserID: 1, Email: "test@gmail.com"},
// 			TestVerification: models.Verification{Email: "test@gmail.com", VerificationCode: "ABCDEF", ExpiresAt: time.Now().Add(time.Hour), AttemptsRemaining: 3},
// 		}},
// 	}

// 	for _, test := range tests {
// 		t.Run(test.desc, func(t *testing.T) {
// 			app := setupApp(&test.db)
// 			path := "/api/auth/verify"
// 			router := gin.New()
// 			router.POST(path, app.VerifyUserHandler)

// 			body := strings.NewReader(test.reqBody)
// 			req := httptest.NewRequest(http.MethodPost, path, body)
// 			w := httptest.NewRecorder()
// 			router.ServeHTTP(w, req)

// 			resp := w.Result()
// 			json, err := io.ReadAll(resp.Body)
// 			if err != nil {
// 				t.Errorf("didn't expect error but got %s", err)
// 			}

// 			assert.Equal(t, test.status, resp.StatusCode)
// 			assert.Equal(t, test.want, string(json))
// 		})
// 	}
// }

// func TestTokenHandler(t *testing.T) {
// 	tests := []struct {
// 		desc    string
// 		reqBody string
// 		status  int
// 		want    string
// 		db      database.MockDBRepo
// 	}{
// 		{desc: "success", reqBody: `{"email": "test@gmail.com", "password": "1234"}`, status: http.StatusOK, want: fmt.Sprintf(`{"status":"success","data":{"token":"%s"}}`, TestToken), db: database.MockDBRepo{
// 			TestUser:         models.User{UserID: 1, Email: "test@gmail.com", Password: "validpass", IsVerified: true},
// 			TestVerification: models.Verification{Email: "test@gmail.com", VerificationCode: "ABCDEF", ExpiresAt: time.Now().Add(time.Hour), AttemptsRemaining: 3},
// 		}},
// 		{desc: "invalid request json body", reqBody: ``, status: http.StatusBadRequest, want: `{"status":"error","message":"unable to parse json body"}`, db: database.MockDBRepo{}},
// 		{desc: "missing parameters", reqBody: `{}`, status: http.StatusBadRequest, want: `{"status":"error","message":"email: required, password: required"}`, db: database.MockDBRepo{}},
// 		{desc: "invalid email", reqBody: `{"email": "invalidemail", "password": "1234"}`, status: http.StatusBadRequest, want: `{"status":"error","message":"email: valid email required"}`, db: database.MockDBRepo{}},
// 		{desc: "user does not exist", reqBody: `{"email": "test@gmail.com", "password": "1234"}`, status: http.StatusUnauthorized, want: `{"status":"error","message":"invalid email or password"}`, db: database.MockDBRepo{
// 			TestUser: models.User{UserID: 0, Email: "test@gmail.com"},
// 		}},
// 		{desc: "user not verified", reqBody: `{"email": "test@gmail.com", "password": "1234"}`, status: http.StatusUnauthorized, want: `{"status":"error","message":"user not verified"}`, db: database.MockDBRepo{
// 			TestUser: models.User{UserID: 1, Email: "test@gmail.com", Password: "validpass", IsVerified: false},
// 		}},
// 		{desc: "invalid password", reqBody: `{"email": "test@gmail.com", "password": "1234"}`, status: http.StatusUnauthorized, want: `{"status":"error","message":"invalid email or password"}`, db: database.MockDBRepo{
// 			TestUser: models.User{UserID: 1, Email: "test@gmail.com", Password: "invalidpass", IsVerified: true},
// 		}},
// 		{desc: "auth token generation failed", reqBody: `{"email": "test@gmail.com", "password": "validpass"}`, status: http.StatusBadRequest, want: `{"status":"error","message":"auth token generation failed"}`, db: database.MockDBRepo{
// 			TestUser: models.User{UserID: 2, Email: "test@gmail.com", Password: "validpass", IsVerified: true},
// 		}},
// 	}

// 	for _, test := range tests {
// 		t.Run(test.desc, func(t *testing.T) {
// 			app := setupApp(&test.db)
// 			path := "/api/auth/token"
// 			router := gin.New()
// 			router.POST(path, app.TokenHandler)

// 			body := strings.NewReader(test.reqBody)
// 			req := httptest.NewRequest(http.MethodPost, path, body)
// 			w := httptest.NewRecorder()
// 			router.ServeHTTP(w, req)

// 			resp := w.Result()
// 			json, err := io.ReadAll(resp.Body)
// 			if err != nil {
// 				t.Errorf("didn't expect error but got %s", err)
// 			}

// 			assert.Equal(t, test.status, resp.StatusCode)
// 			assert.Equal(t, test.want, string(json))
// 		})
// 	}
// }

// func TestDeleteUserHandler(t *testing.T) {
// 	tests := []struct {
// 		desc    string
// 		reqBody string
// 		status  int
// 		want    string
// 		db      database.MockDBRepo
// 	}{
// 		{desc: "success", reqBody: `{"email": "test@gmail.com", "password": "1234"}`, status: http.StatusOK, want: `{"status":"success"}`, db: database.MockDBRepo{}},
// 		{desc: "invalid request json body", reqBody: ``, status: http.StatusBadRequest, want: `{"status":"error","message":"unable to parse json body"}`, db: database.MockDBRepo{}},
// 		{desc: "missing parameters", reqBody: `{}`, status: http.StatusBadRequest, want: `{"status":"error","message":"email: required"}`, db: database.MockDBRepo{}},
// 		{desc: "invalid email", reqBody: `{"email": "invalidemail", "password": "1234"}`, status: http.StatusBadRequest, want: `{"status":"error","message":"email: valid email required"}`, db: database.MockDBRepo{}},
// 		{desc: "delete user failed", reqBody: `{"email": "fail@gmail.com", "password": "1234"}`, status: http.StatusInternalServerError, want: `{"status":"error","message":"DeleteUser failed"}`, db: database.MockDBRepo{}},
// 		{desc: "user not found", reqBody: `{"email": "notfound@gmail.com", "password": "1234"}`, status: http.StatusOK, want: `{"status":"success","data":{"message":"user not found"}}`, db: database.MockDBRepo{}},
// 	}

// 	for _, test := range tests {
// 		t.Run(test.desc, func(t *testing.T) {
// 			app := setupApp(&test.db)
// 			path := "/api/auth/user"
// 			router := gin.New()
// 			router.DELETE(path, app.DeleteUserHandler)

// 			body := strings.NewReader(test.reqBody)
// 			req := httptest.NewRequest(http.MethodDelete, path, body)
// 			w := httptest.NewRecorder()
// 			router.ServeHTTP(w, req)

// 			resp := w.Result()
// 			json, err := io.ReadAll(resp.Body)
// 			if err != nil {
// 				t.Errorf("didn't expect error but got %s", err)
// 			}

// 			assert.Equal(t, test.status, resp.StatusCode)
// 			assert.Equal(t, test.want, string(json))
// 		})
// 	}
// }

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
	default:
		return ""
	}
}

func setupApp(t *testing.T, ctx context.Context) *App {
	// var b bytes.Buffer
	// logWriter := bufio.NewWriter(&b)
	// app := Configs{
	// 	DB:     dbRepo,
	// 	Logger: slog.New(slog.NewJSONHandler(logWriter, nil)),
	// 	Verifier: &MockUserVerifier{
	// 		maxRetries:       3,
	// 		verificationCode: "ABCDEF",
	// 	},
	// 	PasswordEncryptor: &MockPasswordEncryptor{},
	// 	TokenGenerator:    &MockTokenGenerator{},
	// }

	// return &app

	t.Helper()

	pgContainer, err := postgres.Run(
		ctx,
		"postgres:15.3-alpine",
		postgres.WithInitScripts(filepath.Join("..", "..", "migrations", "00000000_000000_init.up.sql")),
		postgres.WithInitScripts(filepath.Join("..", "..", "testdata", "init-db.sql")),
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
		t.Errorf("unexpected error: %s", err)
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

func (t *MockTokenGenerator) GenerateToken(userID int, expiresAtUnixTime int64) (string, error) {
	if userID == 2 {
		return "", errors.New("GenerateToken - unable to generate token")
	}
	return TestToken, nil
}
