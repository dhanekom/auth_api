package main

import (
	"auth_api/internal/models"
	"auth_api/internal/storage"
	"auth_api/internal/storage/database"
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRegisterHandler(t *testing.T) {
	tests := []struct {
		desc    string
		reqBody string
		status  int
		want    string
		db      database.MockDBRepo
	}{
		{desc: "success", reqBody: `{"email": "test@gmail.com", "password": "1234"}`, status: http.StatusOK, want: `{"status":"success","data":{"message":"successfully created user"}}`, db: database.MockDBRepo{}},
		{desc: "invalid request json body", reqBody: ``, status: http.StatusBadRequest, want: `{"status":"error","message":"unable to parse json body"}`, db: database.MockDBRepo{}},
		{desc: "missing parameters", reqBody: `{}`, status: http.StatusBadRequest, want: `{"status":"error","message":"email: required, password: required"}`, db: database.MockDBRepo{}},
		{desc: "invalid email", reqBody: `{"email": "invalidemail", "password": "1234"}`, status: http.StatusBadRequest, want: `{"status":"error","message":"email: valid email required"}`, db: database.MockDBRepo{}},
		{desc: "user already exists", reqBody: `{"email": "test@gmail.com", "password": "1234"}`, status: http.StatusBadRequest, want: `{"status":"error","message":"user already exists"}`, db: database.MockDBRepo{
			TestUsers: []models.User{
				{UserID: 1, Email: "test@gmail.com", Password: "", IsVerified: true},
			},
		}},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			app := setupApp(&test.db)
			path := "/api/auth/register"
			router := gin.New()
			router.POST(path, app.RegisterHandler)

			body := strings.NewReader(test.reqBody)
			req := httptest.NewRequest(http.MethodPost, path, body)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

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
		db      database.MockDBRepo
	}{
		{desc: "success", reqBody: `{"email": "test@gmail.com"}`, status: http.StatusOK, want: `{"status":"success","data":{"verification_code":"ABCDEF"}}`, db: database.MockDBRepo{
			TestUser: models.User{
				UserID:     1,
				Email:      "test@gmail.com",
				Password:   "1234",
				IsVerified: false,
			},
		}},
		{desc: "already verified", reqBody: `{"email": "test@gmail.com"}`, status: http.StatusBadRequest, want: `{"status":"error","message":"user already verified"}`, db: database.MockDBRepo{
			TestUser: models.User{
				UserID:     1,
				Email:      "test@gmail.com",
				Password:   "1234",
				IsVerified: true,
			},
		}},
		{desc: "invalid request json body", reqBody: ``, status: http.StatusBadRequest, want: `{"status":"error","message":"unable to parse json body"}`, db: database.MockDBRepo{}},
		{desc: "missing parameters", reqBody: `{}`, status: http.StatusBadRequest, want: `{"status":"error","message":"email: required"}`, db: database.MockDBRepo{}},
		{desc: "invalid email", reqBody: `{"email": "invalidemail", "password": "1234"}`, status: http.StatusBadRequest, want: `{"status":"error","message":"email: valid email required"}`, db: database.MockDBRepo{}},
		{desc: "user does not exist", reqBody: `{"email": "test@gmail.com"}`, status: http.StatusBadRequest, want: `{"status":"error","message":"user does not exist"}`, db: database.MockDBRepo{
			TestUser: models.User{
				UserID: 0,
			},
		}},
		{desc: "unable to insert verification record", reqBody: `{"email": "fail@gmail.com"}`, status: http.StatusInternalServerError, want: `{"status":"error","message":"InsertOrUpdateVerification failed"}`, db: database.MockDBRepo{
			TestUser: models.User{
				UserID: 1,
			},
		}},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			app := setupApp(&test.db)
			path := "/api/auth/verify"
			router := gin.New()
			router.GET(path, app.GenerateVerificationCodeHandler)

			body := strings.NewReader(test.reqBody)
			req := httptest.NewRequest(http.MethodGet, path, body)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

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
		db      database.MockDBRepo
	}{
		{desc: "success", reqBody: `{"email": "test@gmail.com", "verification_code": "ABCDEF"}`, status: http.StatusOK, want: `{"status":"success"}`, db: database.MockDBRepo{
			TestUser:         models.User{UserID: 1, Email: "test@gmail.com"},
			TestVerification: models.Verification{Email: "test@gmail.com", VerificationCode: "ABCDEF", ExpiresAt: time.Now().Add(time.Hour), AttemptsRemaining: 3},
		}},
		{desc: "invalid request json body", reqBody: ``, status: http.StatusBadRequest, want: `{"status":"error","message":"unable to parse json body"}`, db: database.MockDBRepo{}},
		{desc: "missing parameters", reqBody: `{}`, status: http.StatusBadRequest, want: `{"status":"error","message":"email: required, verification_code: required"}`, db: database.MockDBRepo{}},
		{desc: "invalid email", reqBody: `{"email": "invalidemail", "verification_code": "ABCDEF"}`, status: http.StatusBadRequest, want: `{"status":"error","message":"email: valid email required"}`, db: database.MockDBRepo{}},
		{desc: "user does not exist", reqBody: `{"email": "test@gmail.com", "verification_code": "ABCDEF"}`, status: http.StatusBadRequest, want: `{"status":"error","message":"user does not exist"}`, db: database.MockDBRepo{
			TestUser: models.User{UserID: 0, Email: "test@gmail.com"},
		}},
		{desc: "user does not exist", reqBody: `{"email": "fail@gmail.com", "verification_code": "ABCDEF"}`, status: http.StatusInternalServerError, want: `{"status":"error","message":"no verification data found for user fail@gmail.com"}`, db: database.MockDBRepo{
			TestUser:         models.User{UserID: 1, Email: "fail@gmail.com"},
			TestVerification: models.Verification{Email: "fail@gmail.com"},
		}},
		{desc: "verification code has expired", reqBody: `{"email": "test@gmail.com", "verification_code": "ABCDEF"}`, status: http.StatusBadRequest, want: `{"status":"error","message":"verification code has expired"}`, db: database.MockDBRepo{
			TestUser:         models.User{UserID: 1, Email: "test@gmail.com"},
			TestVerification: models.Verification{Email: "test@gmail.com", VerificationCode: "ABCDEF", ExpiresAt: time.Now().Add(time.Hour * -1), AttemptsRemaining: 3},
		}},
		{desc: "too many attempts", reqBody: `{"email": "test@gmail.com", "verification_code": "ABCDEF"}`, status: http.StatusBadRequest, want: `{"status":"error","message":"verification code has expired"}`, db: database.MockDBRepo{
			TestUser:         models.User{UserID: 1, Email: "test@gmail.com"},
			TestVerification: models.Verification{Email: "test@gmail.com", VerificationCode: "ABCDEF", ExpiresAt: time.Now().Add(time.Hour), AttemptsRemaining: 0},
		}},
		{desc: "invalid verification code", reqBody: `{"email": "test@gmail.com", "verification_code": "INVALID"}`, status: http.StatusBadRequest, want: `{"status":"error","message":"invalid verification code"}`, db: database.MockDBRepo{
			TestUser:         models.User{UserID: 1, Email: "test@gmail.com"},
			TestVerification: models.Verification{Email: "test@gmail.com", VerificationCode: "ABCDEF", ExpiresAt: time.Now().Add(time.Hour), AttemptsRemaining: 3},
		}},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			app := setupApp(&test.db)
			path := "/api/auth/verify"
			router := gin.New()
			router.POST(path, app.VerifyUserHandler)

			body := strings.NewReader(test.reqBody)
			req := httptest.NewRequest(http.MethodPost, path, body)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

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
		db      database.MockDBRepo
	}{
		{desc: "success", reqBody: `{"email": "test@gmail.com", "password": "1234"}`, status: http.StatusOK, want: fmt.Sprintf(`{"status":"success","data":{"token":"%s"}}`, TestToken), db: database.MockDBRepo{
			TestUser:         models.User{UserID: 1, Email: "test@gmail.com", Password: "validpass", IsVerified: true},
			TestVerification: models.Verification{Email: "test@gmail.com", VerificationCode: "ABCDEF", ExpiresAt: time.Now().Add(time.Hour), AttemptsRemaining: 3},
		}},
		{desc: "invalid request json body", reqBody: ``, status: http.StatusBadRequest, want: `{"status":"error","message":"unable to parse json body"}`, db: database.MockDBRepo{}},
		{desc: "missing parameters", reqBody: `{}`, status: http.StatusBadRequest, want: `{"status":"error","message":"email: required, password: required"}`, db: database.MockDBRepo{}},
		{desc: "invalid email", reqBody: `{"email": "invalidemail", "password": "1234"}`, status: http.StatusBadRequest, want: `{"status":"error","message":"email: valid email required"}`, db: database.MockDBRepo{}},
		{desc: "user does not exist", reqBody: `{"email": "test@gmail.com", "password": "1234"}`, status: http.StatusUnauthorized, want: `{"status":"error","message":"invalid email or password"}`, db: database.MockDBRepo{
			TestUser: models.User{UserID: 0, Email: "test@gmail.com"},
		}},
		{desc: "user not verified", reqBody: `{"email": "test@gmail.com", "password": "1234"}`, status: http.StatusUnauthorized, want: `{"status":"error","message":"user not verified"}`, db: database.MockDBRepo{
			TestUser: models.User{UserID: 1, Email: "test@gmail.com", Password: "validpass", IsVerified: false},
		}},
		{desc: "invalid password", reqBody: `{"email": "test@gmail.com", "password": "1234"}`, status: http.StatusUnauthorized, want: `{"status":"error","message":"invalid email or password"}`, db: database.MockDBRepo{
			TestUser: models.User{UserID: 1, Email: "test@gmail.com", Password: "invalidpass", IsVerified: true},
		}},
		{desc: "auth token generation failed", reqBody: `{"email": "test@gmail.com", "password": "validpass"}`, status: http.StatusBadRequest, want: `{"status":"error","message":"auth token generation failed"}`, db: database.MockDBRepo{
			TestUser: models.User{UserID: 2, Email: "test@gmail.com", Password: "validpass", IsVerified: true},
		}},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			app := setupApp(&test.db)
			path := "/api/auth/token"
			router := gin.New()
			router.POST(path, app.TokenHandler)

			body := strings.NewReader(test.reqBody)
			req := httptest.NewRequest(http.MethodPost, path, body)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

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
		db      database.MockDBRepo
	}{
		{desc: "success", reqBody: `{"email": "test@gmail.com", "password": "1234"}`, status: http.StatusOK, want: `{"status":"success"}`, db: database.MockDBRepo{}},
		{desc: "invalid request json body", reqBody: ``, status: http.StatusBadRequest, want: `{"status":"error","message":"unable to parse json body"}`, db: database.MockDBRepo{}},
		{desc: "missing parameters", reqBody: `{}`, status: http.StatusBadRequest, want: `{"status":"error","message":"email: required"}`, db: database.MockDBRepo{}},
		{desc: "invalid email", reqBody: `{"email": "invalidemail", "password": "1234"}`, status: http.StatusBadRequest, want: `{"status":"error","message":"email: valid email required"}`, db: database.MockDBRepo{}},
		{desc: "delete user failed", reqBody: `{"email": "fail@gmail.com", "password": "1234"}`, status: http.StatusInternalServerError, want: `{"status":"error","message":"DeleteUser failed"}`, db: database.MockDBRepo{}},
		{desc: "user not found", reqBody: `{"email": "notfound@gmail.com", "password": "1234"}`, status: http.StatusOK, want: `{"status":"success","data":{"message":"user not found"}}`, db: database.MockDBRepo{}},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			app := setupApp(&test.db)
			path := "/api/auth/user"
			router := gin.New()
			router.DELETE(path, app.DeleteUserHandler)

			body := strings.NewReader(test.reqBody)
			req := httptest.NewRequest(http.MethodDelete, path, body)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

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

func setupApp(dbRepo storage.DBRepo) *Configs {
	var b bytes.Buffer
	logWriter := bufio.NewWriter(&b)
	app := Configs{
		DB:     dbRepo,
		Logger: slog.New(slog.NewJSONHandler(logWriter, nil)),
		Verifier: &MockUserVerifier{
			maxRetries:       3,
			verificationCode: "ABCDEF",
		},
		PasswordEncryptor: &MockPasswordEncryptor{},
		TokenGenerator:    &MockTokenGenerator{},
	}

	return &app
}

type MockUserVerifier struct {
	maxRetries       int
	verificationCode string
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

func (t *MockTokenGenerator) GenerateToken(userID int, expiresAtUnixTime int64) (string, error) {
	if userID == 2 {
		return "", errors.New("GenerateToken - unable to generate token")
	}
	return TestToken, nil
}
