package database

import (
	"auth_api/internal/models"
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

const (
	UserGetSQL = `SELECT user_id, email, password, status, created_at, updated_at
	FROM users
	WHERE email = $1`
	UserCreateSQL = `INSERT INTO users (user_id, email, password, status) values ($1::uuid, $2, $3, $4)`
	UserUpdateSQL = `UPDATE users set email = $1, password = $2, status = $3, updated_at = now() WHERE user_id = $4`
	UserDeleteSQL = `DELETE FROM users where email = $1`

	VerificationUpsertSQL = `INSERT INTO verification (email, verification_type, verification_code, expires_at, attempts_remaining) 
values ($1, $2, $3, $4, $5)
on conflict (email)
  do update set email = $1, verification_type = $2, verification_code = $3, expires_at = $4, attempts_remaining = $5;`
	VerificationGetSQL    = `SELECT email, verification_type, verification_code, expires_at, attempts_remaining, created_at, updated_at FROM verification WHERE email = $1 and verification_type = $2`
	VerificationDeleteSQL = `DELETE FROM verification WHERE email = $1`
)

type PostgresDBRepo struct {
	db *sqlx.DB
}

func NewPostgresDBRepo(db *sqlx.DB) *PostgresDBRepo {
	return &PostgresDBRepo{
		db: db,
	}
}

func (r *PostgresDBRepo) GetUser(ctx context.Context, email string) (*models.User, error) {
	ctxInner, cancel := context.WithTimeout(ctx, time.Second*queryTimeout)
	defer cancel()

	user := models.User{}
	err := r.db.GetContext(ctxInner, &user, UserGetSQL, email)
	if err != nil {
		return nil, fmt.Errorf("unable to get user data: %w", err)
	}

	return &user, nil
}

func (r *PostgresDBRepo) GetUsers(ctx context.Context, email string) ([]models.User, error) {
	ctxInner, cancel := context.WithTimeout(ctx, time.Second*queryTimeout)
	defer cancel()

	users := []models.User{}
	err := r.db.SelectContext(ctxInner, &users, UserGetSQL, email)
	if err != nil {
		return nil, fmt.Errorf("unable to get users data: %w", err)
	}

	return users, nil
}

func (r *PostgresDBRepo) CreateUser(ctx context.Context, user *models.User) error {
	ctxInner, cancel := context.WithTimeout(ctx, time.Second*queryTimeout)
	defer cancel()

	_, err := r.db.ExecContext(ctxInner, UserCreateSQL, user.UserID, user.Email, user.Password, user.Status)
	if err != nil {
		return fmt.Errorf("unable to insert user data: %w", err)
	}

	return nil
}

func (r *PostgresDBRepo) UpdateUser(ctx context.Context, user models.User) error {
	ctxInner, cancel := context.WithTimeout(ctx, time.Second*queryTimeout)
	defer cancel()

	_, err := r.db.ExecContext(ctxInner, UserUpdateSQL, user.Email, user.Password, user.Status, user.UserID)
	if err != nil {
		return fmt.Errorf("unable to update user data: %w", err)
	}

	return nil
}

func (r *PostgresDBRepo) DeleteUser(ctx context.Context, email string) (bool, error) {
	ctxInner, cancel := context.WithTimeout(ctx, time.Second*queryTimeout)
	defer cancel()

	tx := r.db.MustBegin()
	tx.MustExecContext(ctxInner, VerificationDeleteSQL, email)
	result := tx.MustExecContext(ctxInner, UserDeleteSQL, email)
	err := tx.Commit()

	if err != nil {
		return false, fmt.Errorf("unable to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("delete user - unexpected error: %w", err)
	}

	return rowsAffected > 0, nil
}

func (r *PostgresDBRepo) InsertOrUpdateVerification(ctx context.Context, verification models.Verification) error {
	ctxInner, cancel := context.WithTimeout(ctx, time.Second*queryTimeout)
	defer cancel()

	_, err := r.db.ExecContext(ctxInner, VerificationUpsertSQL, verification.Email, verification.VerificationType, verification.VerificationCode, verification.ExpiresAt, verification.AttemptsRemaining)
	if err != nil {
		return fmt.Errorf("unable to insert verification data: %w", err)
	}

	return nil
}

func (r *PostgresDBRepo) GetVerification(ctx context.Context, verificationType string, email string) (*models.Verification, error) {
	ctxInner, cancel := context.WithTimeout(ctx, time.Second*queryTimeout)
	defer cancel()

	var verification models.Verification
	err := r.db.GetContext(ctxInner, &verification, VerificationGetSQL, email, verificationType)
	if err != nil {
		return nil, fmt.Errorf("unable to get verification data: %w", err)
	}

	return &verification, nil
}

func (r *PostgresDBRepo) DeleteVerification(ctx context.Context, email string) error {
	ctxInner, cancel := context.WithTimeout(ctx, time.Second*queryTimeout)
	defer cancel()

	_, err := r.db.ExecContext(ctxInner, VerificationDeleteSQL, email)
	if err != nil {
		return fmt.Errorf("unable to delete verification data: %w", err)
	}

	return nil
}
