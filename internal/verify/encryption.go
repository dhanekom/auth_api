package verify

import (
	"golang.org/x/crypto/bcrypt"
)

type PasswordEncryptor interface {
	GenerateHashedPassword(password string) ([]byte, error)
	CompareHashAndPassword(hashedPassword, password []byte) error
}

type PasswordEncryptorBcrypt struct {
}

func (e PasswordEncryptorBcrypt) GenerateHashedPassword(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), 12)
}

func (e PasswordEncryptorBcrypt) CompareHashAndPassword(hashedPassword, password []byte) error {
	return bcrypt.CompareHashAndPassword(hashedPassword, password)
}
