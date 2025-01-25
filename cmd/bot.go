package main

import (
	"context"
	"fmt"

	"telega_chess/config"
	"telega_chess/internal/db"
	"telega_chess/internal/telegram"
	"telega_chess/internal/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	// cfg
	config.LoadConfig()

	// Инициализация логгера (если нужно):
	utils.InitLogger() // например, зап/лог

	// Инициализация БД
	db.InitDB()

	// Инициализация бота
	botToken := config.Cfg.BotToken // В реальном проекте возьмём из конфига/env
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		utils.Logger.Fatal(fmt.Sprintf("Ошибка при инициализации бота: %v", err))
	}

	telegram.NewHandler(bot)
	// Включим отладочный режим (потом можно отключить)
	bot.Debug = true

	utils.Logger.Info(fmt.Sprintf("Авторизовались как бот: %s", bot.Self.UserName))

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)
	for update := range updates {
		telegram.TelegramHandler.HandleUpdate(context.Background(), update)
	}
}
