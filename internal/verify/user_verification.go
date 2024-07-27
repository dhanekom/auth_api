package verify

import (
	"crypto/rand"
	"errors"
	"io"
	"math"
)

const (
	DefaultMaxCodeLength = 6
	DefaultMaxRetries    = 3
)

type UserVerifier interface {
	Setup(codeLength, maximumRetries int)
	MaxRetries() int
	GenerateVerificationCode() (string, error)
}

type UserVerification struct {
	codeLength     int
	maximumRetries int
}

func (v *UserVerification) Setup(codeLength, maximumRetries int) {
	if codeLength <= 0 {
		codeLength = DefaultMaxCodeLength
	}

	// users.verification_code table column has a max length of 255
	codeLength = int(math.Min(float64(codeLength), 255))

	if maximumRetries <= 0 {
		maximumRetries = DefaultMaxRetries
	}

	v.codeLength = codeLength
	v.maximumRetries = maximumRetries
}

func (v *UserVerification) MaxRetries() int {
	return v.maximumRetries
}

func (v *UserVerification) GenerateVerificationCode() (string, error) {
	max := v.codeLength
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
