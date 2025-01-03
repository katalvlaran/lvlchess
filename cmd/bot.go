package main

import (
	"log"

	"telega_chess/internal/db"
	"telega_chess/internal/telegram"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	// Инициализация БД
	db.InitDB()

	// Инициализация бота
	botToken := "7983098788:AAE7Zgshg3wdhn_L9XMXixAARSTi4Ys8MRw" // В реальном проекте возьмём из конфига/env
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatalf("Ошибка при инициализации бота: %v", err)
	}

	// Включим отладочный режим (потом можно отключить)
	bot.Debug = true

	log.Printf("Авторизовались как бот: %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		// Обработка inline-кнопок
		if update.CallbackQuery != nil {
			telegram.HandleCallback(bot, update.CallbackQuery)
			continue
		}

		// Обработка команд
		if update.Message != nil {
			if update.Message.NewChatMembers != nil {
				telegram.HandleNewChatMembers(bot, update)
				continue
			} else if update.Message.IsCommand() {
				telegram.HandleCommands(bot, update)
			} else {
				// Любые текстовые сообщения
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Используйте /start для списка команд.")
				bot.Send(msg)
			}
		}

		//if update.Message != nil && update.Message.IsCommand() {
		//	telegram.HandleCommands(bot, update)
		//} else if update.Message != nil {
		//	// Любые текстовые сообщения
		//	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Используйте /start для списка команд.")
		//	bot.Send(msg)
		//}
	}
}
