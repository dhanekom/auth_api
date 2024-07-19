package models

import "time"

type Verification struct {
	Email             string    `db:"email"`
	VerificationCode  string    `db:"verification_code"`
	ExpiresAt         time.Time `db:"expires_at"`
	AttemptsRemaining int       `db:"attempts_remaining"`
	CreatedAt         time.Time `db:"created_at"`
	UpdatedAt         time.Time `db:"updated_at"`
}
