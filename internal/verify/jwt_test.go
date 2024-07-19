package verify

import (
	"testing"
	"time"
)

func TestGenerateToken(t *testing.T) {
	tokenGenerator := TokenGeneratorJWT{}
	dateTimeStr := "2999-12-30T01:23:45.678Z"
	myTime, err := time.Parse(time.RFC3339, dateTimeStr)

	if err != nil {
		t.Errorf("unexpected error while parsing date time string: %s", err)
	}

	tokenStr, _ := tokenGenerator.GenerateToken(1, myTime.Unix())

	if err := tokenGenerator.ValidateToken(tokenStr); err != nil {
		t.Errorf("unexpected error while parsing date time string: %v", err.Error())
	}
}
