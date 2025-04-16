package main

import (
	"context"
	"fmt"

	"lvlchess/config"
	"lvlchess/internal/db"
	"lvlchess/internal/telegram"
	"lvlchess/internal/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

/*
Main entry point for lvlChess:
1) Loads configuration from environment (config.LoadConfig).
2) Initializes logger (utils.InitLogger).
3) Initializes DB (db.InitDB).
4) Creates a Telegram Bot API instance using BOT_TOKEN from config.
5) Registers telegram.NewHandler (the main callback structure).
6) Listens for updates in a loop.
*/
func main() {
	// 1) Load environment-based configuration
	config.LoadConfig()

	// 2) Initialize a global logger (Zap in production mode)
	utils.InitLogger() // e.g., logs to stdout, can also set different encoders

	// 3) Initialize a PostgreSQL connection pool + basic schema
	db.InitDB()

	// 4) Build the Telegram bot with the token from config
	botToken := config.Cfg.BotToken
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		// Fatal logs indicate we cannot continue running
		utils.Logger.Fatal(fmt.Sprintf("Failed to initialize the Telegram BotAPI: %v", err))
	}

	// Register the global telegram handler structure
	telegram.NewHandler(bot)

	// (Optional) Debug mode for verbose logging in BotAPI
	bot.Debug = true
	utils.Logger.Info(fmt.Sprintf("Authenticated as bot: %s", bot.Self.UserName))

	// 5) Start receiving updates (long-polling by default).
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	// 6) Process each incoming update in a loop
	for update := range updates {
		// Our handleUpdate is context-based if we need cancellation or deadlines
		telegram.TelegramHandler.HandleUpdate(context.Background(), update)
	}
}
