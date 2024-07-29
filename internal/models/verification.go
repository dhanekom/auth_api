package models

import "time"

const (
	VerificationTypeAccount = "account"
	VerificationTypeReset   = "reset"
)

type Verification struct {
	Email             string    `db:"email"`
	VerificationType  string    `db:"verification_type"`
	VerificationCode  string    `db:"verification_code"`
	ExpiresAt         time.Time `db:"expires_at"`
	AttemptsRemaining int       `db:"attempts_remaining"`
	CreatedAt         time.Time `db:"created_at"`
	UpdatedAt         time.Time `db:"updated_at"`
}
