package worker

import (
	"context"
	"gophemart/internal/app/entity"
	"gophemart/internal/app/repository"
	"gophemart/internal/transport/accrual"
	"gophemart/pkg/logger"
	"time"
)

type OrderProcessor struct {
	orderRepo  repository.OrderRepository
	userRepo   repository.UserRepository
	accrualCli *accrual.Client
}

func NewOrderProcessor(
	orderRepo repository.OrderRepository,
	userRepo repository.UserRepository,
	accrualCli *accrual.Client,
) *OrderProcessor {
	return &OrderProcessor{
		orderRepo:  orderRepo,
		userRepo:   userRepo,
		accrualCli: accrualCli,
	}
}

func (p *OrderProcessor) Run(ctx context.Context, interval time.Duration) {
	logger.Info().
		Dur("interval_seconds", interval).
		Msg("Starting order processor worker")

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info().Msg("Order processor stopped by context")
			return
		case <-ticker.C:
			start := time.Now()
			logger.Debug().Msg("Starting order processing cycle")
			p.processOrders(ctx)
			logger.Debug().
				Dur("duration_ms", time.Since(start)).
				Msg("Order processing cycle completed")
		}
	}
}

func (p *OrderProcessor) processOrders(ctx context.Context) {
	logger.Debug().Msg("Fetching pending orders")

	orders, err := p.orderRepo.FindPending(ctx)
	if err != nil {
		logger.Error().
			Err(err).
			Msg("Failed to get pending orders from repository")
		return
	}

	if len(orders) == 0 {
		logger.Debug().Msg("No pending orders found")
		return
	}

	logger.Info().
		Int("order_count", len(orders)).
		Msg("Processing pending orders")

	for _, order := range orders {
		p.processOrder(ctx, order)
	}
}

func (p *OrderProcessor) processOrder(ctx context.Context, order entity.Order) {
	logger.Debug().
		Str("order_number", order.Number).
		Str("current_status", string(order.Status)).
		Str("user_id", order.UserID).
		Msg("Processing order")

	info, err := p.accrualCli.GetOrderInfo(ctx, order.Number)
	if err != nil {
		if rateLimitErr, ok := err.(*accrual.RateLimitError); ok {
			logger.Warn().
				Err(rateLimitErr).
				Str("order_number", order.Number).
				Dur("retry_after", rateLimitErr.RetryAfter).
				Msg("Accrual service rate limit exceeded, skipping order for now")
			return
		}

		logger.Error().
			Err(err).
			Str("order_number", order.Number).
			Msg("Failed to get order info from accrual service")
		return
	}

	if info == nil {
		logger.Debug().
			Str("order_number", order.Number).
			Msg("Order not found in accrual system")
		return
	}

	newStatus := entity.OrderStatus(info.Status)
	if newStatus == order.Status {
		logger.Debug().
			Str("order_number", order.Number).
			Str("status", string(newStatus)).
			Msg("Order status unchanged, skipping update")
		return
	}

	logger.Info().
		Str("order_number", order.Number).
		Str("old_status", string(order.Status)).
		Str("new_status", string(newStatus)).
		Float64("accrual", info.Accrual).
		Msg("Updating order status")

	err = p.orderRepo.UpdateStatus(ctx, order.Number, newStatus, info.Accrual)
	if err != nil {
		logger.Error().
			Err(err).
			Str("order_number", order.Number).
			Str("new_status", string(newStatus)).
			Msg("Failed to update order status in repository")
		return
	}

	if newStatus == entity.OrderProcessed && info.Accrual > 0 {
		logger.Info().
			Str("order_number", order.Number).
			Str("user_id", order.UserID).
			Float64("accrual", info.Accrual).
			Msg("Adding accrual to user balance")

		err = p.userRepo.AddBalance(ctx, order.UserID, info.Accrual)
		if err != nil {
			logger.Error().
				Err(err).
				Str("order_number", order.Number).
				Str("user_id", order.UserID).
				Float64("accrual", info.Accrual).
				Msg("Failed to add accrual to user balance")
		} else {
			logger.Info().
				Str("order_number", order.Number).
				Str("user_id", order.UserID).
				Float64("accrual", info.Accrual).
				Msg("Accrual successfully added to user balance")
		}
	}

	logger.Info().
		Str("order_number", order.Number).
		Str("new_status", string(newStatus)).
		Msg("Order processing completed")
}
