package accrual

import (
	"context"
	"encoding/json"
	"errors"
	"gophemart/internal/app/entity"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClient_GetOrderInfo(t *testing.T) {
	tests := []struct {
		name           string
		handler        http.HandlerFunc
		expectedResult *OrderInfo
		expectedError  error
	}{
		{
			name: "successful responce",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(OrderInfo{
					Order:   "123",
					Status:  string(entity.OrderProcessed),
					Accrual: 0,
				})
			},
			expectedResult: &OrderInfo{
				Order:   "123",
				Status:  string(entity.OrderProcessed),
				Accrual: 0,
			},
			expectedError: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Запускаем тестовый сервер
			server := httptest.NewServer(tt.handler)
			defer server.Close()

			// Создаем клиент с адресом тестового сервера
			client := NewClient(server.URL)

			// Вызываем тестируемый метод
			result, err := client.GetOrderInfo(context.Background(), "123")

			// Проверяем ошибки
			if tt.expectedError != nil {
				if err == nil {
					t.Fatal("expected error, got nil")
				}

				switch expectedErr := tt.expectedError.(type) {
				case *RateLimitError:
					actualErr, ok := err.(*RateLimitError)
					if !ok {
						t.Fatalf("expected RateLimitError, got %T", err)
					}
					if actualErr.RetryAfter != expectedErr.RetryAfter {
						t.Errorf("expected retry after %v, got %v", expectedErr.RetryAfter, actualErr.RetryAfter)
					}
				default:
					if err.Error() != tt.expectedError.Error() {
						t.Errorf("expected error '%v', got '%v'", tt.expectedError, err)
					}
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}

			// Проверяем результат
			if result != nil || tt.expectedResult != nil {
				if result == nil || tt.expectedResult == nil {
					t.Errorf("result mismatch: expected %v, got %v", tt.expectedResult, result)
				} else if *result != *tt.expectedResult {
					t.Errorf("result mismatch: expected %+v, got %+v", *tt.expectedResult, *result)
				}
			}
		})
	}
}

func TestClient_Timeout(t *testing.T) {
	// Создаем сервер с искусственной задержкой
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(20 * time.Second) // Превышаем таймаут клиента
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL)

	// Создаем контекст с таймаутом (меньшим чем у сервера)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := client.GetOrderInfo(ctx, "123")

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected deadline exceeded error, got: %v", err)
	}
}
