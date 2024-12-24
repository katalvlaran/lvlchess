package game

import (
	"time"

	"github.com/katalvlaran/telega-shess/internal/game/ai"
	"github.com/notnil/chess"
)

// BotHandler обрабатывает игру с ботом
type BotHandler struct {
	bot        *ai.ChessBot
	analyzer   *ai.PositionAnalyzer
	difficulty ai.Difficulty
}

// NewBotHandler создает новый обработчик бота
func NewBotHandler(difficulty ai.Difficulty) *BotHandler {
	return &BotHandler{
		bot:        ai.NewChessBot(difficulty),
		analyzer:   ai.NewPositionAnalyzer(),
		difficulty: difficulty,
	}
}

// MakeBotMove выполняет ход бота
func (bh *BotHandler) MakeBotMove(game *chess.Game) (*chess.Move, error) {
	move := bh.bot.GetMove(game.Position())
	if move == nil {
		return nil, ErrNoValidMoves
	}

	err := game.Move(move)
	if err != nil {
		return nil, err
	}

	return move, nil
}

// AnalyzePosition добавляем метод анализа позиции
func (bh *BotHandler) AnalyzePosition(game *chess.Game) string {
	return bh.analyzer.AnalyzePosition(game.Position())
}

// UpdateGameHandler добавляем метод для обработки игры с ботом
func (gh *GameHandler) HandleBotGame(roomID string, playerID int64) error {
	room, err := gh.roomManager.GetRoom(roomID)
	if err != nil {
		return err
	}

	// Получаем анализ позиции перед ходом бота
	analysis := gh.botHandler.AnalyzePosition(room.State.Game)

	// Делаем ход ботом
	botMove, err := gh.botHandler.MakeBotMove(room.State.Game)
	if err != nil {
		return err
	}

	// Обновляем состояние игры
	room.State.LastMove = time.Now()

	// Генерируем событие хода с анализом
	event := gh.generateEvent(room, botMove, EventMove)
	event.Extra["analysis"] = analysis
	gh.eventHandler.HandleEvent(event)

	return nil
}
