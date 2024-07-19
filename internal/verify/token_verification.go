package verify

import (
	"crypto/rand"
	"errors"
	"io"
	"math"

	"github.com/spf13/viper"
)

const (
	DefaultMaxCodeLength = 6
	DefaultMaxRetries    = 3
)

type UserVerifier interface {
	MaxRetries() int
	GenerateVerificationCode() (string, error)
}

type UserVerification struct {
}

func NewUserVerifier() *UserVerification {
	return &UserVerification{}
}

func (v *UserVerification) maxCodeLength() int {
	max := viper.GetInt("verification.code_length")
	if max == 0 {
		max = DefaultMaxCodeLength
	}

	// users.verification_code table column has a max length of 255
	max = int(math.Min(float64(max), 255))

	return max
}

func (v *UserVerification) MaxRetries() int {
	max := viper.GetInt("verification.max_retries")
	if max == 0 {
		max = DefaultMaxRetries
	}

	return max
}

func (v *UserVerification) GenerateVerificationCode() (string, error) {
	max := v.maxCodeLength()
	buf := make([]byte, max)
	_, err := io.ReadAtLeast(rand.Reader, buf, max)
	if err != nil {
		return "", errors.New("unable to generate verification code")
	}

	for i := 0; i < len(buf); i++ {
		buf[i] = table[int(buf[i])%len(table)]
	}

	return string(buf), nil
}

var table = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
