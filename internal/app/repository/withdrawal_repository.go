package repository

import (
	"context"
	"gophemart/internal/app/entity"
)

type WithdrawalRepository interface {
	Create(ctx context.Context, withdrawal *entity.Withdrawal) error
	FindByUserID(ctx context.Context, userID string) ([]entity.Withdrawal, error)
}
