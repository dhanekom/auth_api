package models

import "time"

type User struct {
	UserID     string    `db:"user_id"`
	Email      string    `db:"email"`
	Password   string    `db:"password"`
	IsVerified bool      `db:"is_verified"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}
