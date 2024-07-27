package verify

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateVerificationCode(t *testing.T) {
	tests := []struct {
		desc               string
		codeLength         int
		maxRetries         int
		expectedCodeLength int
		expectedMaxRetries int
		err                error
	}{
		{desc: "success", codeLength: 6, maxRetries: 3, expectedCodeLength: 6, expectedMaxRetries: 3, err: nil},
		{desc: "long code length", codeLength: 100, maxRetries: 3, expectedCodeLength: 100, expectedMaxRetries: 3, err: nil},
		{desc: "revert to default values", codeLength: -10, maxRetries: 0, expectedCodeLength: 6, expectedMaxRetries: 3, err: nil},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			uv := UserVerification{}
			uv.Setup(test.codeLength, test.maxRetries)

			assert.Equal(t, test.expectedCodeLength, uv.codeLength)
			assert.Equal(t, test.expectedMaxRetries, uv.maximumRetries)

			code, err := uv.GenerateVerificationCode()

			assert.ErrorIs(t, test.err, err)
			assert.Equal(t, len(code), test.expectedCodeLength)
		})
	}
}
