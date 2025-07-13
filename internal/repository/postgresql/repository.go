package postgresql

import (
	"gophemart/internal/app/repository"
	"gorm.io/gorm"
)

type Repository struct {
	User       repository.UserRepository
	Order      repository.OrderRepository
	Withdrawal repository.WithdrawalRepository
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{
		User:  NewUserRepository(db),
		Order: NewOrderRepository(db),
	}
}
