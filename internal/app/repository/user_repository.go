package repository

import (
	"context"
	"errors"
	"gophemart/internal/app/entity"
)

var (
	ErrRocordNotFound = errors.New("ErrRocordNotFound")
)

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	FindByLogin(ctx context.Context, login string) (*entity.User, error)
	FindByID(ctx context.Context, id string) (*entity.User, error)
	UpdateBalance(ctx context.Context, userID string, currentDelta, withdrawnDelta float64) error
	AddBalance(ctx context.Context, userID string, amount float64) error
	CreateWithdrawal(ctx context.Context, withdrawal *entity.Withdrawal) error
}
