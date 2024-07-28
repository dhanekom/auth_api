package storage

import (
	"auth_api/internal/models"
	"context"
)

type DBRepo interface {
	GetUser(ctx context.Context, email string) (*models.User, error)
	GetUsers(ctx context.Context, email string) ([]models.User, error)
	CreateUser(ctx context.Context, user *models.User) error
	UpdateUser(ctx context.Context, user models.User) error
	DeleteUser(ctx context.Context, email string) (bool, error)
	InsertOrUpdateVerification(ctx context.Context, verification models.Verification) error
	GetVerification(ctx context.Context, verificationType string, email string) (*models.Verification, error)
	DeleteVerification(ctx context.Context, email string) error
}
