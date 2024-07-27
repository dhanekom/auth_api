package main

import (
	"auth_api/internal/verify"
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	ctx := context.Background()
	app, err := NewServer(os.Stdout, os.Getenv, "", &verify.UserVerification{}, &verify.PasswordEncryptorBcrypt{}, &verify.JWTTokenUtils{})
	if err != nil {
		fmt.Fprintf(os.Stdout, "%s\n", err)
		os.Exit(1)
	}

	defer app.db.Close()

	if err := run(ctx, app); err != nil {
		fmt.Fprintf(os.Stdout, "%s\n", err)
		os.Exit(1)
	}
}
