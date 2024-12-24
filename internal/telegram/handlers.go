package telegram

import (
	"fmt"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// handleStartCommand обрабатывает команду /start
func (b *Bot) handleStartCommand(message *tgbotapi.Message, session *Session) {
	welcomeText := `Добро пожаловать в шахматного бота! 🎮

Доступные команды:
/create_room - Создать игровую комнату
/play_with_bot - Играть против бота
/move - Сделать ход (например: /move e2-e4)
/draw - Предложить ничью
/surrender - Сдаться

Удачной игры! ♟️`

	b.sendMessage(message.Chat.ID, welcomeText)
}

// handleCreateRoomCommand обрабатывает команду /create_room
func (b *Bot) handleCreateRoomCommand(message *tgbotapi.Message, session *Session) {
	if session.State != StateIdle {
		b.sendMessage(message.Chat.ID, "Вы уже находитесь в игре или создаете ��омнату")
		return
	}

	// Создаем клавиатуру для выбора времени
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("1 минута", "time_1"),
			tgbotapi.NewInlineKeyboardButtonData("5 минут", "time_5"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("30 минут", "time_30"),
			tgbotapi.NewInlineKeyboardButtonData("1 час", "time_60"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Без ограничений", "time_0"),
		),
	)

	b.sessions.UpdateSession(message.From.ID, func(s *Session) {
		s.State = StateSelectingTimeControl
		s.LastMessage = message
	})

	b.sendMessageWithKeyboard(message.Chat.ID, "Выберите контроль времени:", keyboard)
}

// handleCallback обрабатывает нажатия inline-кнопок
func (b *Bot) handleCallback(callback *tgbotapi.CallbackQuery) {
	session := b.sessions.GetSession(callback.From.ID)
	if session == nil {
		return
	}

	data := callback.Data
	if strings.HasPrefix(data, "time_") {
		b.handleTimeControlSelection(callback, session)
		return
	}

	// Обработка других тип��в callback-запросов
	switch data {
	case "accept_draw":
		b.handleDrawAccept(callback, session)
	case "decline_draw":
		b.handleDrawDecline(callback, session)
	}
}

// handleTimeControlSelection обрабатывает выбор времени
func (b *Bot) handleTimeControlSelection(callback *tgbotapi.CallbackQuery, session *Session) {
	timeStr := strings.TrimPrefix(callback.Data, "time_")
	var timeLimit int
	fmt.Sscanf(timeStr, "%d", &timeLimit)

	room, err := b.gameHandler.CreateRoom(callback.From.ID, time.Duration(timeLimit)*time.Minute)
	if err != nil {
		b.sendMessage(callback.Message.Chat.ID, "Ошибка при создании комнаты")
		return
	}

	b.sessions.UpdateSession(callback.From.ID, func(s *Session) {
		s.State = StateWaitingForOpponent
		s.CurrentRoom = room.ID
		s.TimeControl = timeLimit
	})

	shareLink := fmt.Sprintf("t.me/%s?start=%s", callback.Message.Chat.UserName, room.ID)
	response := fmt.Sprintf("Комната создана! Отправьте эту ссылку противнику:\n%s", shareLink)

	b.sendMessage(callback.Message.Chat.ID, response)
}
