package models

import "time"

// OrderResponse представляет модель ответа с информацией о заказе пользователя.
type OrderResponse struct {
	// Number — номер заказа (уникальный идентификатор).
	Number string `json:"number"`
	// Status — статус обработки заказа.
	Status string `json:"status"`
	// Accrual — сумма начисленных баллов за заказ.
	Accrual *float64 `json:"accrual,omitempty"`
	// UploadedAt — дата и время загрузки заказа в систему пользователем.
	UploadedAt time.Time `json:"uploaded_at"`
}

// WithdrawRequest представляет модель запроса на вывод средств
// со счета пользователя в пользу указанного заказа.
type WithdrawRequest struct {
	// Order — номер заказа, в пользу которого производится вывод средств.
	Order string `json:"order"`
	// Sum — сумма для вывода в баллах. Должна быть положительным числом.
	Sum float64 `json:"sum"`
}

// WithdrawalResponse представляет модель ответа с информацией
// о выполненном выводе средств.
type WithdrawalResponse struct {
	// Order — номер заказа, в пользу которого был сделан вывод.
	Order string `json:"order"`
	// Sum — сумма выведенных средств в баллах.
	Sum float64 `json:"sum"`
	// ProcessedAt — дата и время обработки вывода средств.
	ProcessedAt time.Time `json:"processed_at"`
}

// BalanceResponse представляет модель ответа с информацией
// о текущем балансе пользователя.
type BalanceResponse struct {
	// Current — текущий доступный баланс в баллах.
	Current float64 `json:"current"`
	// Withdrawn — сумма всех успешно выведенных средств за все время в баллах
	Withdrawn float64 `json:"withdrawn"`
}
