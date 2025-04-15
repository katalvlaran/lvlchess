package telegram

import (
	"context"
	"fmt"

	// "lvlchess/internal/db" could be used if we needed direct db access here, but we rely on repos in Handler
	"lvlchess/internal/db/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// handleStartCommand is invoked when a user sends the /start command in private chat.
// It registers/updates the user in the DB, checks if user provided something like /start room_<id>,
// and if not, displays a welcome message with inline buttons.
func (h *Handler) handleStartCommand(ctx context.Context, update tgbotapi.Update) {
	// 1) Construct a User object with essential info from Telegram's update.
	p1 := models.User{
		ID:        update.Message.From.ID,
		Username:  update.Message.From.UserName,
		FirstName: update.Message.From.FirstName,
		ChatID:    update.Message.Chat.ID, // For private chat, the chat ID == user ID in Telegram
	}

	// 2) Create or update in DB, ensuring we track them properly.
	h.UserRepo.CreateOrUpdateUser(ctx, &p1)

	// 3) If /start is invoked with "room_<id>", user is joining a specific room.
	args := update.Message.CommandArguments()
	if len(args) > 5 && args[:5] == "room_" {
		roomID := args[5:]
		h.handleJoinRoom(ctx, update, roomID)
		return
	}

	// 4) If not joining a room, present a standard welcome text + inline keyboard menu.
	welcomeText := "Добро пожаловать в Telega-Chess!\n" +
		"Ниже есть несколько возможностей:"

	// We define some inline buttons representing actions (create room, game list, etc.).
	btnCreateRoom := tgbotapi.NewInlineKeyboardButtonData("🆕 Создать комнату", CreateRoom)
	btnMyGames := tgbotapi.NewInlineKeyboardButtonData("📂 Мои игры", GameList)
	btnCreateTournament := tgbotapi.NewInlineKeyboardButtonData("🆕 Создать ТУРНИР", "create_tournament")
	btnMyTournaments := tgbotapi.NewInlineKeyboardButtonData("📃 Мои турниры", "tournament_list")
	btnPlayBot := tgbotapi.NewInlineKeyboardButtonData("🤖 Играть с ботом", PlayWithBot)
	btnSetupRoom := tgbotapi.NewInlineKeyboardButtonData("⚙️ Создать и настроить комнату", SetupRoom)

	// You can arrange these buttons in multiple rows as below.
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(btnCreateRoom, btnMyGames),
		tgbotapi.NewInlineKeyboardRow(btnPlayBot, btnSetupRoom),
		tgbotapi.NewInlineKeyboardRow(btnCreateTournament, btnMyTournaments),
	)

	// 5) Send the message with keyboard attached.
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, welcomeText)
	msg.ReplyMarkup = keyboard
	h.Bot.Send(msg)
}

// handlePlayWithBotCommand is a placeholder for a future feature: playing vs an AI or local engine.
// Currently, we simply send a message "In development."
func (h *Handler) handlePlayWithBotCommand(ctx context.Context, query *tgbotapi.CallbackQuery) {
	msg := tgbotapi.NewMessage(query.Message.Chat.ID, "Игра с ботом в разработке.")
	h.Bot.Send(msg)
}

// handleGameListCommand lists all active rooms (waiting/playing) for the user, if any.
// Called when user presses a "Мои игры" (my games) button, or potentially some command callback.
func (h *Handler) handleGameListCommand(ctx context.Context, query *tgbotapi.CallbackQuery) {
	userID := query.From.ID

	rooms, err := h.RoomRepo.GetPlayingRoomsForUser(ctx, userID)
	if err != nil {
		h.Bot.Send(tgbotapi.NewMessage(query.Message.Chat.ID,
			"Ошибка при получении списка игр: "+err.Error()))
		return
	}

	if len(rooms) == 0 {
		h.Bot.Send(tgbotapi.NewMessage(query.Message.Chat.ID,
			"У вас нет активных игр."))
		return
	}

	// Construct an inline keyboard, one row per room, showing which side is to move.
	var rows [][]tgbotapi.InlineKeyboardButton
	for i, room := range rooms {
		turnTitle := getCurrentTurnUsername(&room)
		buttonText := fmt.Sprintf("Комната_№%d: %s (ход @%s)",
			i+1, room.RoomTitle, turnTitle)
		callbackData := fmt.Sprintf("%s:%s", RoomID, room.RoomID)
		btn := tgbotapi.NewInlineKeyboardButtonData(buttonText, callbackData)
		rows = append(rows, []tgbotapi.InlineKeyboardButton{btn})
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
	msg := tgbotapi.NewMessage(query.Message.Chat.ID, "Ваши активные игры:")
	msg.ReplyMarkup = keyboard
	h.Bot.Send(msg)
}

// getCurrentTurnUsername is a helper that returns who is to move: "белых" or "чёрных."
// In your original code, you might refine it to fetch the actual player's username if WhiteID/BlackID is known.
func getCurrentTurnUsername(r *models.Room) string {
	if r.IsWhiteTurn {
		return "белых"
	}
	return "чёрных"
}
