package http

import (
	"errors"
	"fmt"
	"github.com/labstack/echo"
	"gophemart/internal/app/service"
	"gophemart/internal/handler/http/dto"
	"gophemart/pkg/logger"
	"net/http"
	"strconv"
	"time"
)

type BalanceHandler struct {
	balanceService *service.BalanceService
}

func NewBalanceHandler(balanceService *service.BalanceService) *BalanceHandler {
	return &BalanceHandler{balanceService: balanceService}
}

func (h *BalanceHandler) GetBalance(c echo.Context) error {
	ctx := c.Request().Context()
	userID, ok := c.Get(userIDKey).(string)
	if !ok || userID == "" {
		logger.Error().
			Str("handler", "UploadOrder").
			Msg("UserID not found in context or invalid type")
		return fmt.Errorf("userID not found or invalid type")
	}
	user, err := h.balanceService.GetBalance(ctx, userID)
	if err != nil {
		logger.Error().Str("errerr", err.Error()).Msg("UserID not found in context")

		return c.JSON(http.StatusBadRequest, err)
	}

	response := dto.BalanceResponse{
		Current:   user.CurrentBalance,
		Withdrawn: user.Withdrawn,
	}
	logger.Error().
		Str("user_id", userID).
		Float64("current", user.CurrentBalance).
		Float64("withdrawn", user.Withdrawn).
		Msg("Balance retrieved successfully")

	return c.JSON(http.StatusOK, response)
}

func (h *BalanceHandler) Withdraw(c echo.Context) error {
	userID, ok := c.Get(userIDKey).(string)
	if !ok || userID == "" {
		logger.Error().
			Str("handler", "UploadOrder").
			Msg("UserID not found in context or invalid type")
		return fmt.Errorf("userID not found or invalid type")
	}

	var req dto.WithdrawRequest
	if err := c.Bind(&req); err != nil {
		logger.Warn().
			Err(err).
			Str("handler", "Withdraw").
			Str("user_id", userID).
			Msg("Failed to bind request")
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request format")
	}

	if !isValidLuhn(req.Order) {
		logger.Warn().
			Str("handler", "Withdraw").
			Str("user_id", userID).
			Str("order", req.Order).
			Float64("sum", req.Sum).
			Msg("Invalid order number format")
		return echo.NewHTTPError(http.StatusUnprocessableEntity, "invalid order number")
	}

	ctx := c.Request().Context()

	err := h.balanceService.Withdraw(ctx, userID, req.Order, req.Sum)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInsufficientFunds):
			logger.Warn().
				Str("handler", "Withdraw").
				Str("user_id", userID).
				Str("order", req.Order).
				Float64("sum", req.Sum).
				Msg("Insufficient funds for withdrawal")
			return echo.NewHTTPError(http.StatusPaymentRequired, "insufficient funds")

		case errors.Is(err, service.ErrDuplicateOrder):
			logger.Warn().
				Str("handler", "Withdraw").
				Str("user_id", userID).
				Str("order", req.Order).
				Msg("Order already processed")
			return echo.NewHTTPError(http.StatusConflict, "order already processed")

		case errors.Is(err, service.ErrInvalidOrder):
			logger.Warn().
				Str("handler", "Withdraw").
				Str("user_id", userID).
				Str("order", req.Order).
				Msg("Invalid order number")
			return echo.NewHTTPError(http.StatusUnprocessableEntity, "invalid order number")

		default:
			logger.Error().
				Err(err).
				Str("handler", "Withdraw").
				Str("user_id", userID).
				Str("order", req.Order).
				Float64("sum", req.Sum).
				Msg("Failed to process withdrawal")
			return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
		}
	}

	logger.Info().
		Str("handler", "Withdraw").
		Str("user_id", userID).
		Str("order", req.Order).
		Float64("sum", req.Sum).
		Msg("Withdrawal processed successfully")

	return c.NoContent(http.StatusOK)
}
func isValidLuhn(number string) bool {
	sum := 0
	parity := len(number) % 2

	for i, char := range number {
		digit, err := strconv.Atoi(string(char))
		if err != nil {
			return false
		}

		if i%2 == parity {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
	}

	return sum%10 == 0
}
func (h *BalanceHandler) GetWithdrawals(c echo.Context) error {
	userID, ok := c.Get(userIDKey).(string)
	if !ok || userID == "" {
		logger.Error().Str("handler", "GetWithdrawals").Msg("UserID not found in context")
		return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
	}

	ctx := c.Request().Context()

	withdrawals, err := h.balanceService.GetWithdrawals(ctx, userID)
	if err != nil {
		logger.Error().
			Err(err).
			Str("user_id", userID).
			Str("handler", "GetWithdrawals").
			Msg("Failed to get user withdrawals")
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	if len(withdrawals) == 0 {
		logger.Info().
			Str("user_id", userID).
			Msg("No withdrawals found for user")
		return c.JSON(http.StatusOK, []dto.WithdrawResponce{})
	}

	response := make([]dto.WithdrawResponce, 0, len(withdrawals))
	for _, w := range withdrawals {
		response = append(response, dto.WithdrawResponce{
			Order:       w.OrderNumber,
			Sum:         w.Sum,
			ProcessedAt: w.ProcessedAt.UTC().Format(time.RFC3339),
		})
	}

	logger.Info().
		Str("user_id", userID).
		Int("count", len(withdrawals)).
		Msg("Withdrawals retrieved successfully")

	return c.JSON(http.StatusOK, response)
}
