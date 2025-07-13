package repository

import (
	"context"
	"gophemart/internal/app/entity"
)

type OrderRepository interface {
	Create(ctx context.Context, order *entity.Order) error
	FindByNumber(ctx context.Context, number string) (*entity.Order, error)
	FindByUserID(ctx context.Context, userID string) ([]entity.Order, error)
	Update(ctx context.Context, order *entity.Order) error
	UpdateStatus(ctx context.Context, orderNumber string, status entity.OrderStatus, accrual float64) error
	FindUnprocessed(ctx context.Context) ([]entity.Order, error)
	FindPending(ctx context.Context) ([]entity.Order, error)
	CreateWithdrawal(ctx context.Context, withdrawal *entity.Withdrawal) error

	GetWithdrawalsByUser(ctx context.Context, userID string) ([]entity.Withdrawal, error)
}
