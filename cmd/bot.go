package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"

	"lvlchess/config"
	"lvlchess/internal/db"
	"lvlchess/internal/telegram"
	"lvlchess/internal/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type checkInitDataResponse struct {
	OK       bool   `json:"ok"`
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
}

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

	// Start HTTP‑server for WebApp
	go func() {
		// Раздаём React‑билд по запросам к /telegram-game/
		fs := http.FileServer(http.Dir("frontend/build"))
		http.Handle("/telegram-game/", http.StripPrefix("/telegram-game/", fs))

		// Регистрируем эндпоинт проверки initData
		http.HandleFunc("/api/checkInitData", checkInitDataHandler)

		addr := ":8080" // порт, который вы экспонируете в Dockerfile
		utils.Logger.Info("Starting HTTP server on " + addr)
		if err := http.ListenAndServe(addr, nil); err != nil {
			utils.Logger.Fatal("HTTP server failed: " + err.Error())
		}
	}()

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

func checkInitDataHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "parse form error", http.StatusBadRequest)
		return
	}
	initData := r.PostFormValue("initData")
	if initData == "" {
		http.Error(w, "missing initData", http.StatusBadRequest)
		return
	}

	// 1) разбираем на пары
	parts := strings.Split(initData, "&")
	var hashValue string
	dataPairs := make([]string, 0, len(parts))
	for _, p := range parts {
		if strings.HasPrefix(p, "hash=") {
			hashValue = strings.TrimPrefix(p, "hash=")
		} else {
			dataPairs = append(dataPairs, p)
		}
	}
	if hashValue == "" {
		http.Error(w, "no hash in initData", http.StatusBadRequest)
		return
	}

	// 2) сортируем по ASCII-коду
	sort.Strings(dataPairs)
	dataCheckString := strings.Join(dataPairs, "\n")

	// 3) ключ = SHA256(bot_token)
	botToken := os.Getenv("BOT_TOKEN")
	keyHash := sha256.Sum256([]byte(botToken))

	// 4) HMAC-SHA256
	mac := hmac.New(sha256.New, keyHash[:])
	mac.Write([]byte(dataCheckString))
	expectedMAC := hex.EncodeToString(mac.Sum(nil))

	// 5) безопасное сравнение
	if !hmac.Equal([]byte(expectedMAC), []byte(hashValue)) {
		http.Error(w, "invalid signature", http.StatusForbidden)
		return
	}

	// 6) извлекаем user JSON
	for _, p := range dataPairs {
		if strings.HasPrefix(p, "user=") {
			userJSON, _ := url.QueryUnescape(strings.TrimPrefix(p, "user="))
			var u struct {
				ID        int64  `json:"id"`
				IsBot     bool   `json:"is_bot"`
				FirstName string `json:"first_name"`
				Username  string `json:"username"`
			}
			if err := json.Unmarshal([]byte(userJSON), &u); err != nil {
				http.Error(w, "invalid user json", http.StatusInternalServerError)
				return
			}
			resp := checkInitDataResponse{
				OK:       true,
				UserID:   u.ID,
				Username: u.Username,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}
	}
	http.Error(w, "user not found", http.StatusInternalServerError)
}
