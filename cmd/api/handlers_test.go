package main

import (
	"auth_api/internal/models"
	"auth_api/internal/storage"
	"auth_api/internal/storage/database"
	"bufio"
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
			router := gin.Default()
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

func setupApp(dbRepo storage.DBRepo) *Configs {
	var b bytes.Buffer
	logWriter := bufio.NewWriter(&b)
	app := Configs{
		DB:     dbRepo,
		Logger: slog.New(slog.NewJSONHandler(logWriter, nil)),
	}

	return &app
}
