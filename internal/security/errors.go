package security

import "errors"

var (
	// ErrTokenGeneration возникает при ошибке генерации токена
	ErrTokenGeneration = errors.New("failed to generate secure token")

	// ErrInvalidToken возникает при невалидном токене
	ErrInvalidToken = errors.New("invalid token")

	// ErrTokenExpired возникает при истекшем токене
	ErrTokenExpired = errors.New("token expired")
)
