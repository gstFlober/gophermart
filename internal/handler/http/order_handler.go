package http

import (
	"errors"
	"fmt"
	"github.com/labstack/echo"
	"gophemart/internal/app/service"
	"gophemart/internal/handler/http/dto"
	"gophemart/pkg/logger"
	"io"
	"net/http"
	"strings"
	"time"
)

type OrderHandler struct {
	orderService *service.OrderService
}

func NewOrderHandler(orderService *service.OrderService) *OrderHandler {
	return &OrderHandler{orderService: orderService}
}

func (h *OrderHandler) UploadOrder(c echo.Context) error {
	ctx := c.Request().Context()

	userID, ok := c.Get(userIDKey).(string)
	if !ok || userID == "" {
		logger.Error().
			Str("handler", "OrderHandler.UploadOrder").
			Msg("UserID not found in context or invalid type")
		return fmt.Errorf("userID not found or invalid type")
	}
	body := c.Request().Body
	defer body.Close()

	dataFromBody, err := io.ReadAll(body)
	if err != nil {
		logger.Error().
			Err(err).
			Str("handler", "UploadOrder").
			Msg("Failed to read request body")
		return c.JSON(http.StatusBadRequest, err)
	}

	orderNumber := strings.TrimSpace(string(dataFromBody))
	if orderNumber == "" {
		logger.Warn().
			Str("handler", "UploadOrder").
			Msg("Empty order number provided")

		return c.JSON(http.StatusBadRequest, err)
	}
	if !isValidLuhn(orderNumber) {
		logger.Warn().Str("order_number", orderNumber).Msg("Invalid order number format")
		return echo.NewHTTPError(http.StatusUnprocessableEntity, "invalid order number format")
	}
	err = h.orderService.UploadOrder(ctx, userID, orderNumber)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrOrderAlreadyUploaded):
			logger.Info().
				Str("user_id", userID).
				Str("order_number", orderNumber).
				Msg("Order already uploaded by user")

			return c.NoContent(http.StatusOK)

		case errors.Is(err, service.ErrOrderBelongsToAnotherUser):
			logger.Warn().
				Str("user_id", userID).
				Str("order_number", orderNumber).
				Msg("Order belongs to another user")

			return echo.NewHTTPError(http.StatusConflict, "order already exists for another user")

		default:
			logger.Error().
				Err(err).
				Str("user_id", userID).
				Str("order_number", orderNumber).
				Msg("Failed to upload order")

			return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
		}
	}

	return c.NoContent(http.StatusAccepted)
}

func (h *OrderHandler) GetOrders(c echo.Context) error {
	ctx := c.Request().Context()
	userID, ok := c.Get(userIDKey).(string)
	if !ok || userID == "" {
		logger.Error().
			Str("handler", "OrderHandler.UploadOrder").
			Msg("UserID not found in context or invalid type")
		return fmt.Errorf("userID not found or invalid type")
	}
	orders, err := h.orderService.GetUserOrders(ctx, userID)
	if err != nil {
		logger.Error().
			Err(err).
			Str("user_id", userID).
			Str("handler", "GetOrders").
			Msg(err.Error())

		return c.JSON(http.StatusBadRequest, err)
	}

	responce := make([]dto.OrderResponce, 0, len(orders))
	for _, order := range orders {

		responce = append(responce, dto.OrderResponce{
			Number:     order.Number,
			Status:     string(order.Status),
			Accrual:    order.Accrual,
			UploadedAt: order.UploadedAt.Format(time.RFC3339),
		})
	}

	logger.Info().
		Str("user_id", userID).
		Int("order_count", len(orders)).
		Msg("Orders retrieved successfully")

	return c.JSON(http.StatusOK, responce)
}
