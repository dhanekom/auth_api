package database

import (
	"auth_api/internal/models"
	"context"
	"database/sql"
	"errors"
)

type MockDBRepo struct {
	TestUser         models.User
	TestUsers        []models.User
	TestVerification models.Verification
}

func (r *MockDBRepo) GetUser(ctx context.Context, email string) (*models.User, error) {
	if r.TestUser.UserID == 0 {
		return nil, sql.ErrNoRows
	}

	user := models.User{
		UserID:     r.TestUser.UserID,
		Email:      email,
		Password:   r.TestUser.Password,
		IsVerified: r.TestUser.IsVerified,
	}

	return &user, nil
}

func (r *MockDBRepo) GetUsers(ctx context.Context, email string) ([]models.User, error) {
	users := []models.User{}

	if len(r.TestUsers) > 0 {
		users = append(users, r.TestUsers...)
	}

	return users, nil
}

func (r *MockDBRepo) CreateUser(ctx context.Context, user *models.User) error {
	if r.TestUser.UserID == 2 {
		return errors.New("CreateUser failed")
	}

	user.UserID = r.TestUser.UserID
	user.Email = r.TestUser.Email
	user.Password = r.TestUser.Password
	user.IsVerified = r.TestUser.IsVerified

	return nil
}

func (r *MockDBRepo) UpdateUser(ctx context.Context, user models.User) error {
	if user.UserID == 0 {
		return errors.New("UpdateUser failed")
	}

	return nil
}

func (r *MockDBRepo) DeleteUser(ctx context.Context, email string) (bool, error) {
	if email == "fail@gmail.com" {
		return false, errors.New("DeleteUser failed")
	}
	if email == "notfound@gmail.com" {
		return false, nil
	}
	return true, nil
}

func (r *MockDBRepo) InsertOrUpdateVerification(ctx context.Context, verification models.Verification) error {
	if verification.Email == "fail@gmail.com" {
		return errors.New("InsertOrUpdateVerification failed")
	}
	return nil
}

func (r *MockDBRepo) GetVerification(ctx context.Context, email string) (*models.Verification, error) {
	if r.TestVerification.Email == "fail@gmail.com" {
		return nil, sql.ErrNoRows
	}

	verification := models.Verification{
		Email:             r.TestVerification.Email,
		VerificationCode:  r.TestVerification.VerificationCode,
		ExpiresAt:         r.TestVerification.ExpiresAt,
		AttemptsRemaining: r.TestVerification.AttemptsRemaining,
	}

	return &verification, nil
}

func (r *MockDBRepo) DeleteVerification(ctx context.Context, email string) error {
	return nil
}
