package main

import (
	"auth_api/internal/storage"
	"auth_api/internal/storage/database"
	"auth_api/internal/verify"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/spf13/viper"
)

const webPort = "80"

type Configs struct {
	DB                storage.DBRepo
	Logger            *slog.Logger
	Verifier          verify.UserVerifier
	PasswordEncryptor verify.PasswordEncryptor
	TokenGenerator    verify.TokenGenerator
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// read configs
	viper.SetConfigName("config") // name of config file (without extension)
	viper.SetConfigType("yaml")   // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(".")      // optionally look for config in the working directory
	err := viper.ReadInConfig()   // Find and read the config file
	if err != nil {               // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	host := viper.GetString("db.host")
	port := viper.GetString("db.port")
	dbname := viper.GetString("db.name")
	username := viper.GetString("db.username")
	password := viper.GetString("db.password")

	// connect to DB
	db, err := database.ConnectToPostgres(host, port, dbname, username, password)
	if err != nil {
		log.Fatal(err)
	}

	dbrepo := database.NewPostgresDBRepo(db)
	jwtSecret := viper.GetString("jwt.secret")
	app := Configs{
		DB:                dbrepo,
		Logger:            logger,
		Verifier:          verify.NewUserVerifier(viper.GetInt("verification.code_length"), viper.GetInt("verification.max_retries")),
		PasswordEncryptor: verify.PasswordEncryptorBcrypt{},
		TokenGenerator:    &verify.TokenGeneratorJWT{Secret: jwtSecret},
	}

	srv := http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	// fmt.Printf("Starting servers on port %s\n", webPort)
	app.Logger.Info(fmt.Sprintf("Starting servers on port %s", webPort))

	err = srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
