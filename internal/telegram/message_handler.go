package telegram

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/katalvlaran/telega-shess/internal/game"
)

// handleMessage обрабатывает обычные сообщения
func (b *Bot) handleMessage(message *tgbotapi.Message) {
	session := b.sessions.GetSession(message.From.ID)
	if session == nil {
		return
	}

	switch session.WaitingFor {
	case WaitingForMove:
		b.handleMoveInput(message, session)
	case WaitingForPromotion:
		b.handlePromotionInput(message, session)
	default:
		// Обработка других типов сообщений
		b.handleGenericMessage(message, session)
	}
}

// handleMoveInput обрабатывает ввод хода
func (b *Bot) handleMoveInput(message *tgbotapi.Message, session *Session) {
	moveRegex := regexp.MustCompile(`^([a-h][1-8])-([a-h][1-8])$`)
	if !moveRegex.MatchString(message.Text) {
		b.sendMessage(message.Chat.ID, "Неверный формат хода. Используйте формат: e2-e4")
		return
	}

	parts := strings.Split(message.Text, "-")
	move := game.Move{
		From:      parts[0],
		To:        parts[1],
		Timestamp: time.Now(),
	}

	err := b.gameHandler.MakeMove(session.CurrentRoom, message.From.ID, move)
	if err != nil {
		b.sendMessage(message.Chat.ID, fmt.Sprintf("Ошибка: %v", err))
		return
	}

	// Отправляем обновленную доску
	board, _ := b.gameHandler.GetRenderedBoard(session.CurrentRoom)
	b.sendMessage(message.Chat.ID, board)

	// Обновляем состояние сессии
	session.WaitingFor = WaitingNone
}

// handlePromotionInput обрабатывает выбор фигуры для превращения пешки
func (b *Bot) handlePromotionInput(message *tgbotapi.Message, session *Session) {
	piece := strings.ToUpper(message.Text)
	validPieces := map[string]bool{"Q": true, "R": true, "B": true, "N": true}

	if !validPieces[piece] {
		b.sendMessage(message.Chat.ID, "Выберите фигуру: Q (ферзь), R (ладья), B (слон), N (конь)")
		return
	}

	// Применяем превращение
	move := game.Move{
		From:      session.LastCommand,
		To:        session.LastMessage.Text,
		Promotion: piece,
		Timestamp: time.Now(),
	}

	err := b.gameHandler.MakeMove(session.CurrentRoom, message.From.ID, move)
	if err != nil {
		b.sendMessage(message.Chat.ID, fmt.Sprintf("Ошибка: %v", err))
		return
	}

	board, _ := b.gameHandler.GetRenderedBoard(session.CurrentRoom)
	b.sendMessage(message.Chat.ID, board)

	session.WaitingFor = WaitingNone
}
