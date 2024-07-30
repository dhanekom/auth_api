package models

import "time"

const (
	UserStatusVerifyAccount       = "verify_account"
	UserStatusVerifyResetPassword = "verify_reset"
	UserStatusActive              = "active"
)

type User struct {
	UserID    string    `db:"user_id"`
	Email     string    `db:"email"`
	Password  string    `db:"password"`
	Status    string    `db:"status"`
	Role      string    `db:"role"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (u User) IsVerified() bool {
	return u.Status != UserStatusVerifyAccount
}
