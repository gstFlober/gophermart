package postgresql

import (
	"context"
	"errors"
	"fmt"
	"gophemart/internal/app/entity"
	"gophemart/internal/app/repository"
	"gophemart/pkg/logger"
	"gorm.io/gorm"
	"strings"
)

type OrderRepository struct {
	BaseRepository
}

var (
	ErrNotFound            = errors.New("not found")
	ErrDuplicateKey        = errors.New("duplicate key")
	ErrDuplicateWithdrawal = errors.New("duplicate withdrawal")
)

func NewOrderRepository(db *gorm.DB) repository.OrderRepository {
	return &OrderRepository{BaseRepository{db: db}}
}
func (r *OrderRepository) GetWithdrawalsByUser(
	ctx context.Context,
	userID string,
) ([]entity.Withdrawal, error) {
	logger.Debug().
		Str("method", "OrderRepository.GetWithdrawalsByUser").
		Str("user_id", userID).
		Msg("Fetching user withdrawals")

	var withdrawals []entity.Withdrawal

	result := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("processed_at DESC").
		Find(&withdrawals)

	if result.Error != nil {
		logger.Error().
			Err(result.Error).
			Str("method", "OrderRepository.GetWithdrawalsByUser").
			Str("user_id", userID).
			Msg("Database error when fetching withdrawals")
		return nil, fmt.Errorf("database error: %w", result.Error)
	}

	logger.Debug().
		Str("method", "OrderRepository.GetWithdrawalsByUser").
		Str("user_id", userID).
		Int("count", len(withdrawals)).
		Msg("Successfully retrieved withdrawals")
	return withdrawals, nil
}
func (r *OrderRepository) Create(ctx context.Context, order *entity.Order) error {
	logger.Debug().
		Str("method", "OrderRepository.Create").
		Str("user_id", order.UserID).
		Str("order_number", order.Number).
		Str("status", string(order.Status)).
		Msg("Creating new order")

	err := r.db.WithContext(ctx).Create(order).Error
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			logger.Warn().
				Str("method", "OrderRepository.Create").
				Str("user_id", order.UserID).
				Str("order_number", order.Number).
				Msg("Duplicate order detected")
			return ErrDuplicateKey
		}

		logger.Error().
			Err(err).
			Str("method", "OrderRepository.Create").
			Str("user_id", order.UserID).
			Str("order_number", order.Number).
			Msg("Database error when creating order")
		return fmt.Errorf("database error: %w", err)
	}

	logger.Debug().
		Str("method", "OrderRepository.Create").
		Str("user_id", order.UserID).
		Str("order_number", order.Number).
		Str("status", string(order.Status)).
		Msg("Order created successfully")
	return nil
}
func (r *OrderRepository) FindByNumber(ctx context.Context, number string) (*entity.Order, error) {
	logger.Debug().
		Str("method", "OrderRepository.FindByNumber").
		Str("order_number", number).
		Msg("Finding order by number")

	var order entity.Order
	err := r.db.WithContext(ctx).
		Where("number = ?", number).
		First(&order).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Debug().
				Str("method", "OrderRepository.FindByNumber").
				Str("order_number", number).
				Msg("Order not found")
			return nil, ErrNotFound
		}

		logger.Error().
			Err(err).
			Str("method", "OrderRepository.FindByNumber").
			Str("order_number", number).
			Msg("Database error when finding order")
		return nil, fmt.Errorf("database error: %w", err)
	}

	logger.Debug().
		Str("method", "OrderRepository.FindByNumber").
		Str("order_number", number).
		Str("user_id", order.UserID).
		Str("status", string(order.Status)).
		Msg("Order found successfully")
	return &order, nil
}

func (r *OrderRepository) FindByUserID(ctx context.Context, userID string) ([]entity.Order, error) {
	logger.Debug().
		Str("method", "OrderRepository.FindByUserID").
		Str("user_id", userID).
		Msg("Finding orders by user ID")

	var orders []entity.Order
	result := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&orders)
	if result.Error != nil {
		logger.Error().
			Err(result.Error).
			Str("method", "OrderRepository.FindByUserID").
			Str("user_id", userID).
			Msg("Database error when finding user orders")
		return nil, result.Error
	}

	logger.Debug().
		Str("method", "OrderRepository.FindByUserID").
		Str("user_id", userID).
		Int("count", len(orders)).
		Msg("User orders retrieved successfully")
	return orders, nil
}

func (r *OrderRepository) Update(ctx context.Context, order *entity.Order) error {
	logger.Debug().
		Str("method", "OrderRepository.Update").
		Str("order_number", order.Number).
		Str("status", string(order.Status)).
		Float64("accrual", order.Accrual).
		Msg("Updating order")

	result := r.db.WithContext(ctx).Save(order)
	if result.Error != nil {
		logger.Error().
			Err(result.Error).
			Str("method", "OrderRepository.Update").
			Str("order_number", order.Number).
			Msg("Database error when updating order")
		return result.Error
	}

	logger.Debug().
		Str("method", "OrderRepository.Update").
		Str("order_number", order.Number).
		Int64("rows_affected", result.RowsAffected).
		Msg("Order updated successfully")
	return nil
}

func (r *OrderRepository) FindUnprocessed(ctx context.Context) ([]entity.Order, error) {
	logger.Debug().
		Str("method", "OrderRepository.FindUnprocessed").
		Msg("Finding unprocessed orders")

	var orders []entity.Order
	err := r.db.WithContext(ctx).
		Where("status IN ?", []entity.OrderStatus{entity.OrderNew, entity.OrderProcessing}).
		Find(&orders).Error

	if err != nil {
		logger.Error().
			Err(err).
			Str("method", "OrderRepository.FindUnprocessed").
			Msg("Database error when finding unprocessed orders")
		return nil, err
	}

	logger.Debug().
		Str("method", "OrderRepository.FindUnprocessed").
		Int("count", len(orders)).
		Msg("Unprocessed orders retrieved successfully")
	return orders, nil
}

func (r *OrderRepository) FindPending(ctx context.Context) ([]entity.Order, error) {
	logger.Debug().
		Str("method", "OrderRepository.FindPending").
		Msg("Finding pending orders")

	var orders []entity.Order
	err := r.db.WithContext(ctx).
		Where("status IN ?", []entity.OrderStatus{entity.OrderNew, entity.OrderProcessing}).
		Find(&orders).Error

	if err != nil {
		logger.Error().
			Err(err).
			Str("method", "OrderRepository.FindPending").
			Msg("Database error when finding pending orders")
		return nil, fmt.Errorf("database error: %w", err)
	}

	logger.Debug().
		Str("method", "OrderRepository.FindPending").
		Int("count", len(orders)).
		Msg("Pending orders retrieved successfully")
	return orders, nil
}

func (r *OrderRepository) UpdateStatus(
	ctx context.Context,
	orderNumber string,
	status entity.OrderStatus,
	accrual float64,
) error {
	logger.Debug().
		Str("method", "OrderRepository.UpdateStatus").
		Str("order_number", orderNumber).
		Str("new_status", string(status)).
		Float64("accrual", accrual).
		Msg("Updating order status")

	err := r.db.WithContext(ctx).
		Model(&entity.Order{}).
		Where("number = ?", orderNumber).
		Updates(map[string]interface{}{
			"status":  status,
			"accrual": accrual,
		}).Error

	if err != nil {
		logger.Error().
			Err(err).
			Str("method", "OrderRepository.UpdateStatus").
			Str("order_number", orderNumber).
			Msg("Database error when updating order status")
		return fmt.Errorf("database error: %w", err)
	}

	logger.Debug().
		Str("method", "OrderRepository.UpdateStatus").
		Str("order_number", orderNumber).
		Str("new_status", string(status)).
		Float64("accrual", accrual).
		Msg("Order status updated successfully")
	return nil
}
func (r *OrderRepository) CreateWithdrawal(
	ctx context.Context,
	withdrawal *entity.Withdrawal,
) error {
	logger.Debug().
		Str("method", "OrderRepository.CreateWithdrawal").
		Str("user_id", withdrawal.UserID).
		Str("order_number", withdrawal.OrderNumber).
		Float64("sum", withdrawal.Sum).
		Msg("Creating withdrawal record")

	result := r.db.WithContext(ctx).Create(withdrawal)
	if result.Error != nil {
		if strings.Contains(result.Error.Error(), "duplicate key") {
			logger.Warn().
				Str("method", "OrderRepository.CreateWithdrawal").
				Str("user_id", withdrawal.UserID).
				Str("order_number", withdrawal.OrderNumber).
				Msg("Duplicate withdrawal detected")
			return ErrDuplicateWithdrawal
		}

		logger.Error().
			Err(result.Error).
			Str("method", "OrderRepository.CreateWithdrawal").
			Str("user_id", withdrawal.UserID).
			Str("order_number", withdrawal.OrderNumber).
			Float64("sum", withdrawal.Sum).
			Msg("Database error when creating withdrawal")
		return fmt.Errorf("database error: %w", result.Error)
	}

	logger.Debug().
		Str("method", "OrderRepository.CreateWithdrawal").
		Str("user_id", withdrawal.UserID).
		Str("order_number", withdrawal.OrderNumber).
		Float64("sum", withdrawal.Sum).
		Msg("Withdrawal record created successfully")
	return nil
}
