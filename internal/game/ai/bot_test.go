package ai

import (
	"testing"
	"time"

	"github.com/notnil/chess"
	"github.com/stretchr/testify/assert"
)

func TestChessBot(t *testing.T) {
	t.Run("GetMove returns valid move", func(t *testing.T) {
		bot := NewChessBot(Medium)
		game := chess.NewGame()

		move := bot.GetMove(game.Position())
		assert.NotNil(t, move)

		err := game.Move(move)
		assert.NoError(t, err)
	})

	t.Run("Different difficulty levels", func(t *testing.T) {
		game := chess.NewGame()

		easyBot := NewChessBot(Easy)
		mediumBot := NewChessBot(Medium)
		hardBot := NewChessBot(Hard)

		// Замеряем время для разных уровней сложности
		start := time.Now()
		easyMove := easyBot.GetMove(game.Position())
		easyTime := time.Since(start)

		start = time.Now()
		mediumMove := mediumBot.GetMove(game.Position())
		mediumTime := time.Since(start)

		start = time.Now()
		hardMove := hardBot.GetMove(game.Position())
		hardTime := time.Since(start)

		// Проверяем, что более сложные уровни тратят б��льше времени на анализ
		assert.True(t, mediumTime > easyTime)
		assert.True(t, hardTime > mediumTime)

		// Проверяем, что все ходы валидны
		assert.NotNil(t, easyMove)
		assert.NotNil(t, mediumMove)
		assert.NotNil(t, hardMove)
	})
}
