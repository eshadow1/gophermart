package repository

import "errors"

var (
	ErrInsufficient = errors.New("insufficient funds")
	ErrInvalidOrder = errors.New("invalid order number format")
	ErrOrder        = errors.New("order number not found")
	ErrConflict     = errors.New("order already uploaded by another user")
)
