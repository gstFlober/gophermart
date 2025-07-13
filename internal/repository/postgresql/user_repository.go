package postgresql

import (
	"context"
	"errors"
	"fmt"
	"gophemart/internal/app/entity"
	"gophemart/internal/app/repository"
	"gophemart/pkg/logger"
	"gorm.io/gorm"
)

type UserRepository struct {
	BaseRepository
}

func NewUserRepository(db *gorm.DB) repository.UserRepository {
	return &UserRepository{BaseRepository{db: db}}
}

func (r *UserRepository) Create(ctx context.Context, user *entity.User) error {
	result := r.db.WithContext(ctx).Create(user)
	if result.Error != nil {

		logger.Error().
			Err(result.Error).
			Str("method", "UserRepository.Create").
			Str("login", user.Login).
			Str("user_id", fmt.Sprintf("%d", user.ID)).
			Msg("Failed to create user in database")
		return fmt.Errorf("database error: %w", result.Error)
	}
	logger.Info().
		Str("method", "UserRepository.Create").
		Str("login", user.Login).
		Str("user_id", fmt.Sprintf("%d", user.ID)).
		Msg("User created successfully")
	return nil
}

func (r *UserRepository) FindByLogin(ctx context.Context, login string) (*entity.User, error) {
	var user entity.User
	result := r.db.WithContext(ctx).Where("login = ?", login).First(&user)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			logger.Error().
				Err(repository.ErrRocordNotFound).
				Str("method", "UserRepository.FindByLogin").
				Str("login", login).
				Msg("User not found by login")

			return nil, repository.ErrRocordNotFound
		}
		logger.Error().
			Err(result.Error).
			Str("method", "UserRepository.FindByLogin").
			Str("login", login).
			Msg("Database error when finding user by login")
		return nil, result.Error
	}
	logger.Info().
		Str("method", "UserRepository.FindByLogin").
		Str("login", login).
		Str("user_id", fmt.Sprintf("%d", user.ID)).
		Msg("User found by login")
	return &user, nil
}

func (r *UserRepository) FindByID(ctx context.Context, userID string) (*entity.User, error) {
	logger.Info().
		Str("method", "UserRepository.FindByID").
		Str("user_id", userID).
		Msg("Updating user balance")

	var user entity.User
	result := r.db.WithContext(ctx).Where("ID = ?", userID).First(&user)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			logger.Error().
				Err(ErrNotFound).
				Str("method", "UserRepository.FindByID").
				Str("user_id", userID).
				Msg("User not found by ID")

			return nil, ErrNotFound
		}
		logger.Error().
			Err(result.Error).
			Str("method", "UserRepository.FindByID").
			Str("user_id", userID).
			Msg("Database error when finding user by ID")
		return nil, fmt.Errorf("database error: %w", result.Error)
	}
	logger.Info().
		Str("method", "UserRepository.FindByID").
		Str("user_id", userID).
		Msg("User found by ID")
	return &user, nil
}

func (r *UserRepository) UpdateBalance(ctx context.Context, userID string, newBalance, newWithdrawn float64) error {
	logger.Info().
		Str("method", "UserRepository.UpdateBalance").
		Str("user_id", userID).
		Float64("new_balance", newBalance).
		Float64("new_withdrawn", newWithdrawn).
		Msg("Updating user balance")

	result := r.db.WithContext(ctx).
		Model(&entity.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"current_balance": newBalance,
			"withdrawn":       newWithdrawn,
		})

	if result.Error != nil {
		logger.Error().
			Err(result.Error).
			Str("method", "UserRepository.UpdateBalance").
			Str("user_id", userID).
			Msg("Database error when updating balance")
		return fmt.Errorf("database error: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		logger.Error().
			Str("method", "UserRepository.UpdateBalance").
			Str("user_id", userID).
			Msg("No rows affected when updating balance - user not found")
		return fmt.Errorf("database error: %w", result.Error)
	}
	logger.Info().
		Str("method", "UserRepository.UpdateBalance").
		Str("user_id", userID).
		Int64("rows_affected", result.RowsAffected).
		Msg("User balance updated successfully")
	return nil
}

func (r *UserRepository) AddBalance(ctx context.Context, userID string, amount float64) error {
	logger.Info().
		Str("method", "UserRepository.AddToBalance").
		Str("user_id", userID).
		Float64("amount", amount).
		Msg("Adding to user balance")

	err := r.db.WithContext(ctx).
		Model(&entity.User{}).
		Where("id = ?", userID).
		Update("current_balance", gorm.Expr("current_balance + ?", amount)).Error

	if err != nil {
		logger.Error().
			Err(err).
			Str("method", "UserRepository.AddToBalance").
			Str("user_id", userID).
			Float64("amount", amount).
			Msg("Database error when adding to balance")
		return fmt.Errorf("database error: %w", err)
	}

	logger.Info().
		Str("method", "UserRepository.AddToBalance").
		Str("user_id", userID).
		Float64("amount", amount).
		Msg("Balance added successfully")
	return nil
}

func (r *UserRepository) CreateWithdrawal(ctx context.Context, withdrawal *entity.Withdrawal) error {
	logger.Info().
		Str("method", "UserRepository.CreateWithdrawal").
		Str("user_id", withdrawal.UserID).
		Str("order_number", withdrawal.OrderNumber).
		Float64("sum", withdrawal.Sum).
		Msg("Creating withdrawal record")

	err := r.db.WithContext(ctx).Create(withdrawal).Error
	if err != nil {
		logger.Error().
			Err(err).
			Str("method", "UserRepository.CreateWithdrawal").
			Str("user_id", withdrawal.UserID).
			Str("order_number", withdrawal.OrderNumber).
			Float64("sum", withdrawal.Sum).
			Msg("Database error when creating withdrawal")
		return fmt.Errorf("database error: %w", err)
	}

	logger.Info().
		Str("method", "UserRepository.CreateWithdrawal").
		Str("user_id", withdrawal.UserID).
		Str("order_number", withdrawal.OrderNumber).
		Float64("sum", withdrawal.Sum).
		Msg("Withdrawal record created successfully")
	return nil
}
