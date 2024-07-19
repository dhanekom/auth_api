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

func MaxCodeLength() int {
	max := viper.GetInt("verification.code_length")
	if max == 0 {
		max = DefaultMaxCodeLength
	}

	// users.verification_code table column has a max length of 255
	max = int(math.Min(float64(max), 255))

	return max
}

func MaxRetries() int {
	max := viper.GetInt("verification.max_retries")
	if max == 0 {
		max = DefaultMaxRetries
	}

	return max
}

func GenerateVerificationCode(max int) (string, error) {
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
