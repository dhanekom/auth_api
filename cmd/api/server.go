package main

import (
	"auth_api/internal/storage"
	"auth_api/internal/storage/database"
	"auth_api/internal/verify"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
)

type Configs struct {
	DB                storage.DBRepo
	Logger            *slog.Logger
	Verifier          verify.UserVerifier
	PasswordEncryptor verify.PasswordEncryptor
	TokenUtils        verify.TokenUtils
	UserTokenSecret   string
	AdminTokenSecret  string
}

type App struct {
	server  *http.Server
	configs *Configs
	db      *sqlx.DB
}

func NewServer(w io.Writer, getenv func(string) string, dbConnStr string, verifier verify.UserVerifier, passwordEncryptor verify.PasswordEncryptor, TokenUtils verify.TokenUtils) (*App, error) {
	logger := slog.New(slog.NewJSONHandler(w, nil))

	EnvReader := NewEnvReader(getenv)

	hostAddr := EnvReader.GetString("AUTH_HOST_ADDR")
	hostPort := EnvReader.GetString("AUTH_HOST_PORT", "80")

	dbConnectionStr := dbConnStr
	if dbConnectionStr == "" {
		dbConnectionStr = EnvReader.GetString("AUTH_DB_CONNECTION_STRING")
	}

	jwtSecret := EnvReader.GetString("AUTH_JWT_SECRET")
	if jwtSecret == "" {
		return nil, errors.New("AUTH_JWT_SECRET environment variable requires a value")
	}

	userTokenSecret := EnvReader.GetString("AUTH_USER_TOKEN_SECRET")
	if userTokenSecret == "" {
		return nil, errors.New("AUTH_USER_TOKEN_SECRET environment variable requires a value")
	}

	adminTokenSecret := EnvReader.GetString("AUTH_ADMIN_TOKEN_SECRET")
	if userTokenSecret == "" {
		return nil, errors.New("AUTH_ADMIN_TOKEN_SECRET environment variable requires a value")
	}

	verificationCodeLength := EnvReader.GetInt("AUTH_VERIFICATION_CODE_LENGTH", 6)
	verificationMaxRetries := EnvReader.GetInt("AUTH_VERIFICATION_MAX_RETRIES", 6)

	// connect to DB
	db, err := database.ConnectToPostgres(dbConnectionStr)
	if err != nil {
		return nil, err
	}
	// Don't close the connect here. It will be done later
	// defer db.Close()

	verifier.Setup(verificationCodeLength, verificationMaxRetries)
	TokenUtils.Setup(jwtSecret)

	var dbrepo storage.DBRepo = database.NewPostgresDBRepo(db)
	configs := Configs{
		DB:                dbrepo,
		Logger:            logger,
		Verifier:          verifier,
		PasswordEncryptor: passwordEncryptor,
		TokenUtils:        TokenUtils,
		UserTokenSecret:   userTokenSecret,
		AdminTokenSecret:  adminTokenSecret,
	}

	srv := http.Server{
		Addr:    net.JoinHostPort(hostAddr, hostPort),
		Handler: configs.routes(),
	}

	return &App{
		server:  &srv,
		configs: &configs,
		db:      db,
	}, nil
}

func run(ctx context.Context, app *App) error {
	go func() {
		app.configs.Logger.Info(fmt.Sprintf("Starting servers on %s", app.server.Addr))

		if err := app.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "error listening and serving: %s\n", err)
		}
	}()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		if err := app.server.Shutdown(shutdownCtx); err != nil {
			fmt.Fprintf(os.Stderr, "error shutting down http server: %s\n", err)
		}

	}()
	wg.Wait()

	return nil
}
