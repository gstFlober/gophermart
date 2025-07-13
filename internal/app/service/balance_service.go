package service

import (
	"context"
	"errors"
	"fmt"
	"gophemart/internal/app/entity"
	"gophemart/internal/app/repository"
	"gophemart/internal/repository/postgresql"
	"gophemart/pkg/logger"
	"math"
	"time"
)

type BalanceService struct {
	userRepo       repository.UserRepository
	orderRepo      repository.OrderRepository
	withdrawalRepo repository.WithdrawalRepository
}

var (
	ErrInsufficientFunds  = errors.New("insufficient funds")
	ErrInvalidOrderNumber = errors.New("invalid order number")
	ErrInvalidOrder       = errors.New("invalid order")
	ErrUserNotFound       = errors.New("user not found")
)

func NewBalanceService(
	userRepo repository.UserRepository,
	orderRepo repository.OrderRepository,
	withdrawalRepo repository.WithdrawalRepository) *BalanceService {
	return &BalanceService{
		userRepo:       userRepo,
		orderRepo:      orderRepo,
		withdrawalRepo: withdrawalRepo,
	}
}

func (s *BalanceService) GetWithdrawals(
	ctx context.Context,
	userID string,
) ([]entity.Withdrawal, error) {
	logger.Info().
		Str("method", "GetWithdrawals").
		Str("userID", userID)
	withdrawals, err := s.orderRepo.GetWithdrawalsByUser(ctx, userID)
	if err != nil {
		logger.Error().
			Err(err).
			Str("userID", userID).
			Msg("failed to get withdrawals")
		return nil, err
	}

	return withdrawals, nil
}
func (s *BalanceService) GetBalance(ctx context.Context, userID string) (*entity.User, error) {
	logger.Info().
		Str("method", "GetBalance").
		Str("user_id", userID).
		Msg("Fetching user balance")

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, postgresql.ErrNotFound) {
			logger.Warn().
				Str("user_id", userID).
				Msg("User not found when fetching balance")
			return nil, ErrUserNotFound
		}
		logger.Error().
			Err(err).
			Str("user_id", userID).
			Msg("Database error when fetching balance")
		return nil, fmt.Errorf("database error: %w", err)
	}
	logger.Info().
		Str("user_id", userID).
		Float64("current_balance", user.CurrentBalance).
		Float64("withdrawn", user.Withdrawn).
		Msg("Successfully retrieved user balance")

	return user, nil
}

func (s *BalanceService) GetWithdtawals(ctx context.Context, userID string) ([]entity.Withdrawal, error) {
	return s.withdrawalRepo.FindByUserID(ctx, userID)
}

func (s *BalanceService) Withdraw(ctx context.Context, userID, orderNumber string, sum float64) error {
	logger.Info().
		Str("method", "Withdraw").
		Str("user_id", userID).
		Str("order_number", orderNumber).
		Float64("sum", sum).
		Msg("Processing withdrawal request")

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		logger.Error().
			Err(err).
			Str("user_id", userID).
			Str("method", "Withdraw").
			Msg("Failed to find user")
		return fmt.Errorf("failed to find user: %w", err)
	}

	if user.CurrentBalance < sum {
		logger.Warn().
			Str("user_id", userID).
			Float64("current_balance", user.CurrentBalance).
			Float64("requested_sum", sum).
			Msg("Insufficient funds for withdrawal")
		return ErrInsufficientFunds
	}

	newBalance := math.Round((user.CurrentBalance-sum)*100) / 100
	newWithdrawn := math.Round((user.Withdrawn+sum)*100) / 100

	err = s.userRepo.UpdateBalance(ctx, userID, newBalance, newWithdrawn)
	if err != nil {
		logger.Error().
			Err(err).
			Str("user_id", userID).
			Str("order", orderNumber).
			Float64("sum", sum).
			Msg("Failed to update user balance")
		return fmt.Errorf("failed to update balance: %w", err)
	}

	withdrawal := &entity.Withdrawal{
		UserID:      userID,
		OrderNumber: orderNumber,
		Sum:         sum,
		ProcessedAt: time.Now().UTC(),
	}

	err = s.orderRepo.CreateWithdrawal(ctx, withdrawal)
	if err != nil {
		logger.Error().
			Err(err).
			Str("user_id", userID).
			Str("order", orderNumber).
			Float64("sum", sum).
			Msg("Failed to create withdrawal record")
		return fmt.Errorf("failed to create withdrawal: %w", err)
	}
	logger.Info().
		Str("user_id", userID).
		Str("order_number", orderNumber).
		Float64("sum", sum).
		Float64("new_balance", newBalance).
		Msg("Withdrawal processed successfully")
	return nil
}
