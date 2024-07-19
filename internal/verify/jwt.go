package verify

import (
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

type TokenGenerator interface {
	GenerateToken(userID int, expiresAtUnixTime int64) (string, error)
}

type TokenGeneratorJWT struct {
	Secret string
}

func (t *TokenGeneratorJWT) GenerateToken(userID int, expiresAtUnixTime int64) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"exp": expiresAtUnixTime,
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString([]byte(t.Secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (t *TokenGeneratorJWT) ValidateToken(tokenStr string) error {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(t.Secret), nil
	})

	if err != nil {
		return fmt.Errorf("token validation failed: %w", err)
	}

	if !token.Valid {
		return errors.New("invalid token")
	}

	return nil
}
