package telegram

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/katalvlaran/telega-shess/internal/game"
	"github.com/katalvlaran/telega-shess/internal/utils"
)

// NewBot создает нового бота
func NewBot(token string, debug bool) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %v", err)
	}

	api.Debug = debug

	bot := &Bot{
		api:         api,
		gameHandler: game.NewGameHandler(),
		log:         utils.Logger(),
		sessions:    NewSessionManager(),
	}

	return bot, nil
}

// Start запускает бота
func (b *Bot) Start() error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)
	b.updates = updates

	b.log.Info("Bot started")

	for update := range updates {
		go b.handleUpdate(update)
	}

	return nil
}

// handleUpdate обрабатывает обновления от Telegram
func (b *Bot) handleUpdate(update tgbotapi.Update) {
	defer func() {
		if r := recover(); r != nil {
			b.log.WithField("panic", r).Error("Recovered from panic in update handler")
		}
	}()

	// Обработка команд
	if update.Message != nil && update.Message.IsCommand() {
		b.handleCommand(update.Message)
		return
	}

	// Обработка callback-запросов (inline-кнопки)
	if update.CallbackQuery != nil {
		b.handleCallback(update.CallbackQuery)
		return
	}

	// Обработка обычных сообщений
	if update.Message != nil {
		b.handleMessage(update.Message)
		return
	}
}

// handleCommand обрабатывает команды
func (b *Bot) handleCommand(message *tgbotapi.Message) {
	session := b.sessions.GetSession(message.From.ID)
	if session == nil {
		session = b.sessions.CreateSession(message.From.ID)
	}

	switch message.Command() {
	case utils.CmdStart:
		b.handleStartCommand(message, session)
	case utils.CmdCreateRoom:
		b.handleCreateRoomCommand(message, session)
	case utils.CmdPlayBot:
		b.handlePlayBotCommand(message, session)
	case utils.CmdMove:
		b.handleMoveCommand(message, session)
	case utils.CmdDraw:
		b.handleDrawCommand(message, session)
	case utils.CmdSurrender:
		b.handleSurrenderCommand(message, session)
	case utils.CmdBless:
		b.handleBlessCommand(message, session)
	default:
		b.sendMessage(message.Chat.ID, "Неизвестная команда")
	}
}

// sendMessage отправляет сообщение в чат
func (b *Bot) sendMessage(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := b.api.Send(msg)
	return err
}

// sendMessageWithKeyboard отправляет сообщение с клавиатурой
func (b *Bot) sendMessageWithKeyboard(chatID int64, text string, keyboard interface{}) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard
	_, err := b.api.Send(msg)
	return err
}
