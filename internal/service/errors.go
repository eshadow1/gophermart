package service

import "errors"

var (
	ErrEmptyBody      = errors.New("empty order number")
	ErrValidationLuhn = errors.New("invalid order number format")
	ErrInsufficient   = errors.New("insufficient funds")
	ErrOrderNotFound  = errors.New("order not found")
)
