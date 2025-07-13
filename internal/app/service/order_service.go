package service

import (
	"context"
	"errors"
	"fmt"
	"gophemart/internal/app/entity"
	"gophemart/internal/app/repository"
	"gophemart/internal/repository/postgresql"
	"gophemart/internal/transport/accrual"
	"gophemart/pkg/logger"
	"time"
)

type OrderService struct {
	orderRepo     repository.OrderRepository
	userRepo      repository.UserRepository
	accrualClient *accrual.Client
}

var (
	ErrInvalidInput              = errors.New("invalid input")
	ErrOrderAlreadyUploaded      = errors.New("order already uploaded by user")
	ErrOrderBelongsToAnotherUser = errors.New("order belongs to another user")
	ErrOrderAlreadyExists        = errors.New("order already exists")
	ErrDuplicateOrder            = errors.New("order already exists")
)

func NewOrderService(orderRepo repository.OrderRepository, userRepo repository.UserRepository, accrualClient *accrual.Client) *OrderService {
	return &OrderService{
		orderRepo:     orderRepo,
		userRepo:      userRepo,
		accrualClient: accrualClient,
	}
}

func (s *OrderService) UploadOrder(ctx context.Context, userID string, number string) error {
	logger.Info().
		Str("method", "UploadOrder").
		Str("user_id", userID).
		Str("order_number", number).
		Msg("Processing order upload")

	existingOrder, err := s.orderRepo.FindByNumber(ctx, number)
	if err != nil && !errors.Is(err, postgresql.ErrNotFound) {
		logger.Error().
			Err(err).
			Str("user_id", userID).
			Str("order_number", number).
			Msg("Database error when checking existing order")
		return fmt.Errorf("database error: %w", err)
	}

	if existingOrder != nil {
		if existingOrder.UserID == userID {
			logger.Info().
				Str("user_id", userID).
				Str("order_number", number).
				Msg("Order already uploaded by same user")
			return ErrOrderAlreadyUploaded
		}
		logger.Warn().
			Str("user_id", userID).
			Str("order_number", number).
			Str("order_owner", existingOrder.UserID).
			Msg("Order belongs to another user")
		return ErrOrderBelongsToAnotherUser
	}

	newOrder := &entity.Order{
		Number:    number,
		UserID:    userID,
		Status:    "NEW",
		CreatedAt: time.Now(),
	}

	if err := s.orderRepo.Create(ctx, newOrder); err != nil {
		if errors.Is(err, postgresql.ErrDuplicateKey) {
			logger.Warn().
				Str("user_id", userID).
				Str("order_number", number).
				Msg("Order already exists (race condition detected)")
			return s.UploadOrder(ctx, userID, number)
		}
		logger.Error().
			Err(err).
			Str("user_id", userID).
			Str("order_number", number).
			Msg("Failed to create order in database")
		return fmt.Errorf("failed to create order: %w", err)
	}
	logger.Info().
		Str("user_id", userID).
		Str("order_number", number).
		Str("order_status", "NEW").
		Msg("Order successfully uploaded")
	return nil

}

func (s *OrderService) GetUserOrders(ctx context.Context, userID string) ([]entity.Order, error) {

	logger.Info().
		Str("method", "GetUserOrders").
		Str("user_id", userID).
		Msg("Fetching user orders")

	orders, err := s.orderRepo.FindByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrRocordNotFound) {
			logger.Info().
				Str("user_id", userID).
				Msg("No orders found for user")
			return []entity.Order{}, nil
		}

		logger.Error().
			Err(err).
			Str("user_id", userID).
			Str("method", "GetUserOrders").
			Msg("Failed to retrieve user orders")

		return nil, fmt.Errorf("failed to get user orders: %w", err)
	}

	logger.Info().
		Str("user_id", userID).
		Int("order_count", len(orders)).
		Msg("Successfully retrieved user orders")

	return orders, nil
}
