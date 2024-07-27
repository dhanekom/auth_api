package verify

import (
	"testing"
)

func TestGenerateToken(t *testing.T) {
	tokenGenerator := JWTTokenUtils{}
	tokenGenerator.Setup("secret")

	tokenStr, _ := tokenGenerator.GenerateToken(1, 24)

	if err := tokenGenerator.ValidateToken(tokenStr); err != nil {
		t.Errorf("unexpected error while validating token: %v", err.Error())
	}
}
