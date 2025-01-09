package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	//"github.com/username/telega-chess/internal/utils"
)

// HandleUpdate - универсальная точка входа
// HandleUpdate - универсальная точка входа
func HandleUpdate(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	switch {
	case update.Message != nil:
		handleMessage(bot, update)
	case update.CallbackQuery != nil:
		handleCallback(bot, update.CallbackQuery)
	case update.MyChatMember != nil:
		//handleContinueSetup(bot, update.CallbackQuery)
	case update.ChatMember != nil:
		HandleNewChatMembers(bot, update)
	}
}

// handleMessage - обрабатываем сообщения/команды
func handleMessage(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	msg := update.Message

	if msg.NewChatMembers != nil {
		HandleNewChatMembers(bot, update)
	}

	// Проверка: если в группе и msg.IsCommand():
	if msg.Chat.IsGroup() || msg.Chat.IsSuperGroup() {
		if msg.IsCommand() {
			if msg.Command() == "setroom" {
				handleSetRoomCommand(bot, update)
			} else {
				// Отключаем остальные команды
				reply := tgbotapi.NewMessage(msg.Chat.ID, "Здесь команды не работают. Используйте /setroom <room_id> или кнопки.")
				bot.Send(reply)
			}
		} else {
			// Любой текст -> "Используйте кнопки..."
			//reply := tgbotapi.NewMessage(msg.Chat.ID, "Используйте кнопки (inline).")
			reply := tgbotapi.NewMessage(msg.Chat.ID, "🌚")
			bot.Send(reply)
		}
		return
	}

	// Если это личка
	if msg.IsCommand() {
		switch msg.Command() {
		case "start":
			handleStartCommand(bot, update) // в logic.go
		case "create_room":
			handleCreateRoomCommand(bot, update)
		case "play_with_bot":
			handlePlayWithBotCommand(bot, update)
		case "game_list":
			handleGameListCommand(bot, update)
		default:
			bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Команда не распознана. Используйте кнопки или /start."))
		}
	} else {
		bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "🌚"))
	}
}
