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

	// 2) Проверка, если /start room_... (старый сценарий handleJoinRoom)
	args := update.Message.CommandArguments()
	if len(args) > 5 && args[:5] == "room_" {
		roomID := args[5:]
		handleJoinRoom(bot, update, roomID)
		return
	}

	// 3) Вывод приветствия (можно чуть скорректировать текст)
	welcomeText := "Добро пожаловать в Telega-Chess!\n" +
		"Ниже есть несколько возможностей:"

	// 4) Формируем Inline-кнопки (4 штуки)
	//    a) «🆕 Создать комнату»
	//    b) «📂 Мои игры»
	//    c) «🤖 Играть с ботом»
	//    d) «⚙️ Создать и настроить комнату»
	btnCreateRoom := tgbotapi.NewInlineKeyboardButtonData("🆕 Создать комнату", "create_room")
	btnMyGames := tgbotapi.NewInlineKeyboardButtonData("📂 Мои игры", "game_list")
	btnPlayBot := tgbotapi.NewInlineKeyboardButtonData("🤖 Играть с ботом", "play_with_bot")
	btnSetupRoom := tgbotapi.NewInlineKeyboardButtonData("⚙️ Создать и настроить комнату", "setup_room")

	// собираем одну строку/несколько, как удобнее
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(btnCreateRoom, btnMyGames),
		tgbotapi.NewInlineKeyboardRow(btnPlayBot, btnSetupRoom),
	)
	// 5) Отправляем сообщение + inline-клавиатуру
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, welcomeText)
	msg.ReplyMarkup = keyboard
	bot.Send(msg)
}

func handlePlayWithBotCommand(bot *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery) {
	msg := tgbotapi.NewMessage(query.Message.Chat.ID, "Игра с ботом в разработке.")
	bot.Send(msg)
}

func handleGameListCommand(bot *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery) {
	msg := tgbotapi.NewMessage(query.Message.Chat.ID, "Ваши активные игры (заглушка).\n1. Комната 12345.\n2. Комната 67890.")
	bot.Send(msg)
}
