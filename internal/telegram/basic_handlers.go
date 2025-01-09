package telegram

import (
	"telega_chess/internal/db"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func handleStartCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update) {

	// 1) Сохраним пользователя
	p1 := db.User{
		ID:        update.Message.From.ID,
		Username:  update.Message.From.UserName,
		FirstName: update.Message.From.FirstName,
		ChatID:    update.Message.Chat.ID, // Личная переписка
	}
	db.CreateOrUpdateUser(&p1)

	args := update.Message.CommandArguments() // то, что идёт после /start
	if len(args) > 5 && args[:5] == "room_" {
		roomID := args[5:]
		handleJoinRoom(bot, update, roomID)
		return
	}

	// Стандартное приветствие, если нет room_
	messageText := "Добро пожаловать в Telega-Chess!\n" +
		"Команды:\n" +
		"- /create_room — создать новую игровую комнату.\n" +
		"- /game_list — вернуться к текущей игре.\n" +
		"- /play_with_bot — играть против AI (заглушка)."
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, messageText)
	bot.Send(msg)
}

func handlePlayWithBotCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Игра с ботом в разработке.")
	bot.Send(msg)
}

func handleGameListCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ваши активные игры (заглушка).\n1. Комната 12345.\n2. Комната 67890.")
	bot.Send(msg)
}
