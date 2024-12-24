package utils

import (
	"os"
	"strconv"
	"time"
)

// Шахматные фигуры Unicode
const (
	KingWhite   = "♔"
	QueenWhite  = "♕"
	RookWhite   = "♖"
	BishopWhite = "♗"
	KnightWhite = "♘"
	PawnWhite   = "♙"

	KingBlack   = "♚"
	QueenBlack  = "♛"
	RookBlack   = "♜"
	BishopBlack = "♝"
	KnightBlack = "♞"
	PawnBlack   = "♟"

	EmptySquare = "."
)

// Специальные символы
const (
	CaptureSymbol   = "💥"
	ForkSymbol      = "🔌"
	PromotionSymbol = "🪄"
	SparklesSymbol  = "✨"
)

// Команды бота
const (
	CmdStart      = "start"
	CmdCreateRoom = "create_room"
	CmdPlayBot    = "play_with_bot"
	CmdMove       = "move"
	CmdDraw       = "draw"
	CmdSurrender  = "surrender"
	CmdBless      = "bless"
)

// Game states
const (
	GameStateNew      = "new"
	GameStateActive   = "active"
	GameStateFinished = "finished"
)

// Event types
const (
	EventMove    = "move"
	EventGameEnd = "game_end"
	EventDraw    = "draw"
)

// Cache settings
const (
	DefaultCacheSize     = 10000
	CacheCleanupInterval = 1 * time.Hour
	CacheEntryLifetime   = 24 * time.Hour
)

// Security settings
const (
	MaxTokens       = 1000
	TokenExpiration = 24 * time.Hour
	RateLimitPerMin = 60
)

// Bot settings
const (
	DefaultDifficulty = "medium"
	MaxGames          = 100
	MoveTimeout       = 30 * time.Second
)

// Configuration структура для конфигурации приложения
type Config struct {
	BotToken    string
	Difficulty  string
	MaxGames    int
	CacheSize   int
	Environment string
}

// LoadConfig загружает конфигурацию из переменных окружения
func LoadConfig() *Config {
	return &Config{
		BotToken:    os.Getenv("BOT_TOKEN"),
		Difficulty:  getEnvOrDefault("DIFFICULTY", DefaultDifficulty),
		MaxGames:    getEnvAsIntOrDefault("MAX_GAMES", MaxGames),
		CacheSize:   getEnvAsIntOrDefault("CACHE_SIZE", DefaultCacheSize),
		Environment: getEnvOrDefault("ENV", "development"),
	}
}

// getEnvOrDefault возвращает значение переменной окружения или значение по умолчанию
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsIntOrDefault возвращает числовое значение переменной окружения или значение по умолчанию
func getEnvAsIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// Logger settings
const (
	LogTimeFormat = "2006-01-02 15:04:05"
	LogFilePath   = "logs/chessbot.log"
)

// Error messages
const (
	ErrInvalidMove       = "Invalid move"
	ErrGameNotFound      = "Game not found"
	ErrInvalidGameState  = "Invalid game state"
	ErrUnauthorized      = "Unauthorized access"
	ErrRateLimitExceeded = "Rate limit exceeded"
)

// Success messages
const (
	MsgGameCreated     = "Game created successfully"
	MsgMoveMade        = "Move made successfully"
	MsgGameFinished    = "Game finished"
	MsgDrawOffered     = "Draw offered"
	MsgDrawAccepted    = "Draw accepted"
	MsgDrawRejected    = "Draw rejected"
	MsgGameSurrendered = "Game surrendered"
)
