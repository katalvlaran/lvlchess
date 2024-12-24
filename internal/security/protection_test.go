package security

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenManager(t *testing.T) {
	t.Run("token creation and validation", func(t *testing.T) {
		tm := NewTokenManager(10)
		require.NotNil(t, tm)

		// Создаем токен
		token, err := tm.CreateToken(123)
		require.NoError(t, err)
		require.NotEmpty(t, token)

		// Проверяем токен
		userID, valid := tm.ValidateToken(token)
		assert.True(t, valid)
		assert.Equal(t, int64(123), userID)
	})

	t.Run("invalid token validation", func(t *testing.T) {
		tm := NewTokenManager(10)
		require.NotNil(t, tm)

		// Проверяем пустой токен
		userID, valid := tm.ValidateToken("")
		assert.False(t, valid)
		assert.Equal(t, int64(0), userID)

		// Проверяем несуществующий токен
		userID, valid = tm.ValidateToken("invalid_token")
		assert.False(t, valid)
		assert.Equal(t, int64(0), userID)
	})

	t.Run("token expiration", func(t *testing.T) {
		tm := NewTokenManager(10)
		require.NotNil(t, tm)

		token, err := tm.CreateToken(123)
		require.NoError(t, err)

		// Модифицируем время истечения
		tm.tokens[token] = TokenInfo{
			UserID:    123,
			ExpiresAt: time.Now().Add(-time.Hour),
			LastUsed:  time.Now(),
		}

		// Проверяем истекший токен
		userID, valid := tm.ValidateToken(token)
		assert.False(t, valid)
		assert.Equal(t, int64(0), userID)
	})

	t.Run("max tokens limit", func(t *testing.T) {
		tm := NewTokenManager(2)
		require.NotNil(t, tm)

		// Создаем максимальное количество токенов
		token1, err := tm.CreateToken(1)
		require.NoError(t, err)
		token2, err := tm.CreateToken(2)
		require.NoError(t, err)

		// Создаем еще один токен
		token3, err := tm.CreateToken(3)
		require.NoError(t, err)

		// Проверяем, что старый токен удален
		_, valid := tm.ValidateToken(token1)
		assert.False(t, valid)

		// Проверяем, что новые токены действительны
		_, valid = tm.ValidateToken(token2)
		assert.True(t, valid)
		_, valid = tm.ValidateToken(token3)
		assert.True(t, valid)
	})

	t.Run("cleanup expired tokens", func(t *testing.T) {
		tm := NewTokenManager(10)
		require.NotNil(t, tm)

		token, err := tm.CreateToken(123)
		require.NoError(t, err)

		// Модифицируем время истечения
		tm.tokens[token] = TokenInfo{
			UserID:    123,
			ExpiresAt: time.Now().Add(-2 * time.Hour),
			LastUsed:  time.Now().Add(-2 * time.Hour),
		}

		// Запускаем очистку
		tm.cleanup()

		// Проверяем, что токен удален
		_, valid := tm.ValidateToken(token)
		assert.False(t, valid)
	})
}
