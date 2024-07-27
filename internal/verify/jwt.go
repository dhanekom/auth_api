package verify

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenUtils interface {
	Setup(secret string)
	GenerateToken(userID int, hours int) (string, error)
}

type JWTTokenUtils struct {
	secret string
}

func (t *JWTTokenUtils) Setup(secret string) {
	t.secret = secret
}

func (t *JWTTokenUtils) GenerateToken(userID int, hours int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(time.Hour * time.Duration(hours)).Unix(),
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString([]byte(t.secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (t *JWTTokenUtils) ValidateToken(tokenStr string) error {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(t.secret), nil
	})

	if err != nil {
		return fmt.Errorf("token validation failed: %w", err)
	}

	if !token.Valid {
		return errors.New("invalid token")
	}

	return nil
}
