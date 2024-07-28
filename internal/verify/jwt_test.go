package verify

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateToken(t *testing.T) {
	tokenGenerator := JWTTokenUtils{}
	tokenGenerator.Setup("secret")

	tokenStr, _ := tokenGenerator.GenerateToken("0460d39a-9c81-48bd-86ed-7154f44ac611", 24)

	if err := tokenGenerator.ValidateToken(tokenStr); err != nil {
		t.Errorf("unexpected error while validating token: %v", err.Error())
	}
}

func TestValidateToken(t *testing.T) {
	tokenGenerator := JWTTokenUtils{}
	tokenGenerator.Setup("secret")

	err := tokenGenerator.ValidateToken("")
	assert.Error(t, err)

	err = tokenGenerator.ValidateToken("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c")
	assert.Error(t, err)
}
