package telegram

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// handleMoveCommand обрабатывает команду /move
func (b *Bot) handleMoveCommand(message *tgbotapi.Message, session *Session) {
	if session.State != StateInGame {
		b.sendMessage(message.Chat.ID, "Вы не находитесь в игре")
		return
	}

	// Создаем клавиатуру с возможными ходами
	keyboard := b.generateMoveKeyboard(session)
	b.sendMessageWithKeyboard(message.Chat.ID, "Выберите ход:", keyboard)
}

// handleDrawCommand обрабатывает команду /draw
func (b *Bot) handleDrawCommand(message *tgbotapi.Message, session *Session) {
	if session.State != StateInGame {
		b.sendMessage(message.Chat.ID, "Вы не находитесь в игре")
		return
	}

	// Создаем клавиатуру для принятия/отклонения ничьей
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Принять", "accept_draw"),
			tgbotapi.NewInlineKeyboardButtonData("Отклонить", "decline_draw"),
		),
	)

	// Отправляем сообщение противнику
	opponentID := session.OpponentID
	b.sendMessageWithKeyboard(opponentID, "Противник предлагает ничью", keyboard)
	b.sendMessage(message.Chat.ID, "Предложение ничьей отправлено")
}

// handleSurrenderCommand обрабатывает команду /surrender
func (b *Bot) handleSurrenderCommand(message *tgbotapi.Message, session *Session) {
	if session.State != StateInGame {
		b.sendMessage(message.Chat.ID, "Вы не находитесь в игре")
		return
	}

	err := b.gameHandler.ResignGame(session.CurrentRoom, message.From.ID)
	if err != nil {
		b.sendMessage(message.Chat.ID, fmt.Sprintf("Ошибка: %v", err))
		return
	}

	// Отправляем сообщение обоим игрокам
	b.sendMessage(message.Chat.ID, "Вы сдались")
	b.sendMessage(session.OpponentID, "Противник сдался! Вы победили!")

	// Сбрасываем состояние сессий
	b.resetGameSessions(session)
}

// handlePlayBotCommand обрабатывает команду /play_with_bot
func (b *Bot) handlePlayBotCommand(message *tgbotapi.Message, session *Session) {
	if session.State != StateIdle {
		b.sendMessage(message.Chat.ID, "Вы уже находитесь в игре")
		return
	}

	// Создаем комнату для игры с ботом
	room, err := b.gameHandler.CreateRoom(message.From.ID, 0) // Без ограничения времени
	if err != nil {
		b.sendMessage(message.Chat.ID, "Ошибка при создании игры")
		return
	}

	b.sessions.UpdateSession(message.From.ID, func(s *Session) {
		s.State = StateInGame
		s.CurrentRoom = room.ID
		s.OpponentID = -1 // Специальный ID для бота
	})

	board, _ := b.gameHandler.GetRenderedBoard(room.ID)
	b.sendMessage(message.Chat.ID, "Игра с ботом начата! Вы играете белыми.\n"+board)
}
