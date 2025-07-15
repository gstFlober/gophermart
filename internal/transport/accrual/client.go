package accrual

import (
	"context"
	"encoding/json"
	"fmt"
	"gophemart/pkg/logger"
	"io"
	"net/http"
	"strconv"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type OrderInfo struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual,omitempty"`
}

func (c *Client) GetOrderInfo(ctx context.Context, orderNumber string) (*OrderInfo, error) {
	url := fmt.Sprintf("%s/api/orders/%s", c.baseURL, orderNumber)

	logger.Debug().
		Str("method", "Client.GetOrderInfo").
		Str("url", url).
		Str("order_number", orderNumber).
		Msg("Sending request to accrual service")

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		logger.Error().
			Err(err).
			Str("order_number", orderNumber).
			Msg("Failed to create request to accrual service")
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	start := time.Now()
	resp, err := c.httpClient.Do(req)
	duration := time.Since(start)

	if err != nil {
		logger.Error().
			Err(err).
			Str("order_number", orderNumber).
			Dur("duration_ms", duration).
			Msg("Request to accrual service failed")
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	logger.Debug().
		Str("method", "Client.GetOrderInfo").
		Str("order_number", orderNumber).
		Int("status_code", resp.StatusCode).
		Dur("duration_ms", duration).
		Msg("Received response from accrual service")

	switch resp.StatusCode {
	case http.StatusOK:
		var info OrderInfo
		if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
			logger.Error().
				Err(err).
				Str("order_number", orderNumber).
				Msg("Failed to decode response from accrual service")
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}

		logger.Debug().
			Str("order_number", orderNumber).
			Str("status", info.Status).
			Float64("accrual", info.Accrual).
			Msg("Successfully retrieved order info")
		return &info, nil

	case http.StatusNoContent:
		logger.Debug().
			Str("order_number", orderNumber).
			Msg("Order not found in accrual system (204 No Content)")
		return nil, nil

	case http.StatusTooManyRequests:
		retryAfterStr := resp.Header.Get("Retry-After")
		retryAfter, err := strconv.Atoi(retryAfterStr)
		if err != nil || retryAfter <= 0 {
			retryAfter = 60
			logger.Warn().
				Str("retry_after_header", retryAfterStr).
				Str("order_number", orderNumber).
				Msg("Invalid Retry-After header, using default 60 seconds")
		} else {
			logger.Debug().
				Str("order_number", orderNumber).
				Int("retry_after", retryAfter).
				Msg("Parsed Retry-After header")
		}

		retryDuration := time.Duration(retryAfter) * time.Second
		logger.Warn().
			Str("order_number", orderNumber).
			Dur("retry_after", retryDuration).
			Msg("Rate limit exceeded in accrual service")

		return nil, &RateLimitError{
			RetryAfter: retryDuration,
			Message:    "too many requests to accrual service",
		}

	default:
		body, _ := io.ReadAll(resp.Body)
		bodyStr := string(body)
		if len(bodyStr) > 1024 {
			bodyStr = bodyStr[:1024] + "..."
		}

		logger.Error().
			Str("order_number", orderNumber).
			Int("status_code", resp.StatusCode).
			Str("response_body", bodyStr).
			Msg("Unexpected status code from accrual service")

		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, bodyStr)
	}
}

type RateLimitError struct {
	RetryAfter time.Duration
	Message    string
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("%s (retry after %v)", e.Message, e.RetryAfter)
}
