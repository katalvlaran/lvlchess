package main

import (
	"fmt"

	"telega_chess/internal/db"
	"telega_chess/internal/telegram"
	"telega_chess/internal/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {

	// 1) Инициализация логгера (если нужно):
	utils.InitLogger() // например, зап/лог

	// Инициализация БД
	db.InitDB()

	// Инициализация бота
	botToken := "7983098788:AAE7Zgshg3wdhn_L9XMXixAARSTi4Ys8MRw" // В реальном проекте возьмём из конфига/env
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		utils.Logger.Fatal(fmt.Sprintf("Ошибка при инициализации бота: %v", err))
	}

	// Включим отладочный режим (потом можно отключить)
	bot.Debug = true

	utils.Logger.Info(fmt.Sprintf("Авторизовались как бот: %s", bot.Self.UserName))

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)
	for update := range updates {
		telegram.HandleUpdate(bot, update)

		/*!!!!!!!!!!!!!!*/
		/*
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
			}*/
	}
}
