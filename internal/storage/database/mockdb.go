package database

import (
	"auth_api/internal/models"
	"context"
	"errors"
)

type MockDBRepo struct {
	TestUser         models.User
	TestUsers        []models.User
	TestVerification models.Verification
}

func (r *MockDBRepo) GetUser(ctx context.Context, email string) (*models.User, error) {
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
	if r.TestUser.UserID == 99 {
		return errors.New("CreateUser failed")
	}

	user.UserID = r.TestUser.UserID
	user.Email = r.TestUser.Email
	user.Password = r.TestUser.Password
	user.IsVerified = r.TestUser.IsVerified

	return nil
}

func (r *MockDBRepo) UpdateUser(ctx context.Context, user models.User) error {
	if user.UserID == 99 {
		return errors.New("UpdateUser failed")
	}

	return nil
}

func (r *MockDBRepo) DeleteUser(ctx context.Context, email string) (bool, error) {
	return true, nil
}

func (r *MockDBRepo) InsertOrUpdateVerification(ctx context.Context, verification models.Verification) error {
	return nil
}

func (r *MockDBRepo) GetVerification(ctx context.Context, email string) (*models.Verification, error) {
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
