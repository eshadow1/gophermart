// Package client предоставляет HTTP-клиент для взаимодействия с внешней
// системой начислений.
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

// AccrualClient реализует HTTP-клиент для взаимодействия с внешней
// системой начислений.
type AccrualClient struct {
	baseURL string
	client  *http.Client
}

// AccrualResponse представляет модель ответа от внешней системы начислений.
// Содержит информацию о статусе обработки заказа и сумме начисления.
type AccrualResponse struct {
	// Order — номер заказа.
	Order string `json:"order"`
	// Status — статус обработки заказа
	Status string `json:"status"`
	// Accrual — сумма начисленных баллов.
	Accrual float64 `json:"accrual,omitempty"`
}

// RateLimitError представляет специализированную ошибку, возвращаемую
// при получении HTTP 429 (Too Many Requests) от внешней системы.
type RateLimitError struct {
	RetryAfter time.Duration
}

// Error возвращает текстовое представление ошибки rate limiting.
func (e *RateLimitError) Error() string {
	return fmt.Sprintf("rate limited: retry after %s", e.RetryAfter)
}

// NewAccrualClient создает и возвращает новый экземпляр HTTP-клиента
// для взаимодействия с внешней системой начислений.
func NewAccrualClient(baseURL string) *AccrualClient {
	return &AccrualClient{
		baseURL: baseURL,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

// GetAccrual запрашивает информацию о начислении для указанного номера заказа
// во внешней системе начислений.
func (c *AccrualClient) GetAccrual(ctx context.Context, orderNum string) (*AccrualResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/api/orders/%s", c.baseURL, orderNum), http.NoBody)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var res AccrualResponse
		if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
			return nil, fmt.Errorf("decode response: %w", err)
		}
		return &res, nil
	case http.StatusNoContent:
		return nil, nil
	case http.StatusTooManyRequests:
		retrySec, _ := strconv.Atoi(resp.Header.Get("Retry-After"))
		if retrySec <= 0 {
			retrySec = 60
		}
		return nil, &RateLimitError{RetryAfter: time.Duration(retrySec) * time.Second}
	default:
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil, fmt.Errorf("external service error: %d", resp.StatusCode)
	}
}
