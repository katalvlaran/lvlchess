package telegram

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// generateMoveKeyboard создает клавиатуру с возможными ходами
func (b *Bot) generateMoveKeyboard(session *Session) tgbotapi.InlineKeyboardMarkup {
	// Получаем возможные ходы
	validMoves := b.gameHandler.GetValidMoves(session.CurrentRoom, session.UserID)

	var rows [][]tgbotapi.InlineKeyboardButton
	currentRow := []tgbotapi.InlineKeyboardButton{}

	for _, move := range validMoves {
		moveStr := fmt.Sprintf("%s-%s", move.From, move.To)
		button := tgbotapi.NewInlineKeyboardButtonData(moveStr, "move_"+moveStr)

		currentRow = append(currentRow, button)
		if len(currentRow) == 3 { // По 3 кнопки в ряд
			rows = append(rows, currentRow)
			currentRow = []tgbotapi.InlineKeyboardButton{}
		}
	}

	if len(currentRow) > 0 {
		rows = append(rows, currentRow)
	}

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// resetGameSessions сбрасывает состояние игровых сессий
func (b *Bot) resetGameSessions(session *Session) {
	// Сбрасываем сессию текущего игрока
	b.sessions.UpdateSession(session.UserID, func(s *Session) {
		s.State = StateIdle
		s.CurrentRoom = ""
		s.OpponentID = 0
		s.WaitingFor = WaitingNone
	})

	// Сбрасываем сессию противника
	if session.OpponentID > 0 {
		b.sessions.UpdateSession(session.OpponentID, func(s *Session) {
			s.State = StateIdle
			s.CurrentRoom = ""
			s.OpponentID = 0
			s.WaitingFor = WaitingNone
		})
	}
}
